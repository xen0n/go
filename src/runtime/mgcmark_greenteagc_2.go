// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Green Tea mark algorithm
//
// The core idea behind Green Tea is simple: achieve better locality during
// mark/scan by delaying scanning so that we can accumulate objects to scan
// within the same span, then scan the objects that have accumulated on the
// span all together.
//
// By batching objects this way, we increase the chance that adjacent objects
// will be accessed, amortize the cost of accessing object metadata, and create
// better opportunities for prefetching. We can take this even further and
// optimize the scan loop by size class (not yet completed) all the way to the
// point of applying SIMD techniques to really tear through the heap.
//
// Naturally, this depends on being able to create opportunties to batch objects
// together. The basic idea here is to have two sets of mark bits. One set is the
// regular set of mark bits ("marks"), while the other essentially says that the
// objects have been scanned already ("scans"). When we see a pointer for the first
// time we set its mark and enqueue its span. We track these spans in work queues
// with a FIFO policy, unlike workbufs which have a LIFO policy. Empirically, a
// FIFO policy appears to work best for accumulating objects to scan on a span.
// Later, when we dequeue the span, we find both the union and intersection of the
// mark and scan bitsets. The union is then written back into the scan bits, while
// the intersection is used to decide which objects need scanning, such that the GC
// is still precise.
//
// Below is the bulk of the implementation, focusing on the worst case
// for locality, small objects. Specifically, those that are smaller than
// a few cache lines in size and whose metadata is stored the same way (at the
// end of the span).

//go:build !goexperiment.greenteagc

package runtime

import (
	"internal/goarch"
	"internal/runtime/atomic"
	"internal/runtime/gc"
	"internal/runtime/sys"
	"unsafe"
)

// spanQueue11 is a P-local stealable span queue.
type spanQueue11 struct {
	// head, tail, and ring represent a local non-thread-safe ring buffer.
	head, tail uint32
	ring       [256]objptr11

	// putsSinceDrain counts the number of put calls since the last drain.
	putsSinceDrain int

	// chain contains state visible to other Ps.
	//
	// In particular, that means a linked chain of single-producer multi-consumer
	// ring buffers where the single producer is this P only.
	//
	// This linked chain structure is based off the sync.Pool dequeue.
	chain struct {
		// head is the spanSPMC to put to. This is only accessed
		// by the producer, so doesn't need to be synchronized.
		head *spanSPMC11

		// tail is the spanSPMC to steal from. This is accessed
		// by consumers, so reads and writes must be atomic.
		tail atomic.UnsafePointer // *spanSPMC
	}
}

// putFast tries to put s onto the queue, but may fail if it's full.
func (q *spanQueue11) putFast(s objptr11) (ok bool) {
	if q.tail-q.head == uint32(len(q.ring)) {
		return false
	}
	q.ring[q.tail%uint32(len(q.ring))] = s
	q.tail++
	return true
}

// put puts s onto the queue.
//
// Returns whether the caller should spin up a new worker.
func (q *spanQueue11) put(s objptr11) bool {
	// The constants below define the period of and volume of
	// spans we spill to the spmc chain when the local queue is
	// not full.
	//
	// spillPeriod must be > spillMax, otherwise that sets the
	// effective maximum size of our local span queue. Even if
	// we have a span ring of size N, but we flush K spans every
	// K puts, then K becomes our effective maximum length. When
	// spillPeriod > spillMax, then we're always spilling spans
	// at a slower rate than we're accumulating them.
	const (
		// spillPeriod defines how often to check if we should
		// spill some spans, counted in the number of calls to put.
		spillPeriod = 64

		// spillMax defines, at most, how many spans to drain with
		// each spill.
		spillMax = 16
	)

	if q.putFast(s) {
		// Occasionally try to spill some work to generate parallelism.
		q.putsSinceDrain++
		if q.putsSinceDrain >= spillPeriod {
			// Reset even if we don't drain, so we don't check every time.
			q.putsSinceDrain = 0

			// Try to drain some spans. Don't bother if there's very
			// few of them or there's already spans in the spmc chain.
			n := min((q.tail-q.head)/2, spillMax)
			if n > 4 && q.chainEmpty() {
				q.drain(n)
				return true
			}
		}
		return false
	}

	// We're out of space. Drain out our local spans.
	q.drain(uint32(len(q.ring)) / 2)
	if !q.putFast(s) {
		throw("failed putFast after drain")
	}
	return true
}

// flush publishes all spans in the local queue to the spmc chain.
func (q *spanQueue11) flush() {
	n := q.tail - q.head
	if n == 0 {
		return
	}
	q.drain(n)
}

// empty returns true if there's no more work on the queue.
//
// Not thread-safe. Must only be called by the owner of q.
func (q *spanQueue11) empty() bool {
	// Check the local queue for work.
	if q.tail-q.head > 0 {
		return false
	}
	return q.chainEmpty()
}

// chainEmpty returns true if the spmc chain is empty.
//
// Thread-safe.
func (q *spanQueue11) chainEmpty() bool {
	// Check the rest of the rings for work.
	r := (*spanSPMC11)(q.chain.tail.Load())
	for r != nil {
		if !r.empty() {
			return false
		}
		r = (*spanSPMC11)(r.prev.Load())
	}
	return true
}

// drain publishes n spans from the local queue to the spmc chain.
func (q *spanQueue11) drain(n uint32) {
	q.putsSinceDrain = 0

	if q.chain.head == nil {
		// N.B. We target 1024, but this may be bigger if the physical
		// page size is bigger, or if we can fit more uintptrs into a
		// physical page. See newSpanSPMC docs.
		r := newSpanSPMC11(1024)
		q.chain.head = r
		q.chain.tail.StoreNoWB(unsafe.Pointer(r))
	}

	// Try to drain some of the queue to the head spmc.
	if q.tryDrain(q.chain.head, n) {
		return
	}
	// No space. Create a bigger spmc and add it to the chain.

	// Double the size of the next one, up to a maximum.
	//
	// We double each time so we can avoid taking this slow path
	// in the future, which involves a global lock. Ideally we want
	// to hit a steady-state where the deepest any queue goes during
	// a mark phase can fit in the ring.
	//
	// However, we still set a maximum on this. We set the maximum
	// to something large to amortize the cost of lock acquisition, but
	// still at a reasonable size for big heaps and/or a lot of Ps (which
	// tend to be correlated).
	//
	// It's not too bad to burn relatively large-but-fixed amounts of per-P
	// memory if we need to deal with really, really deep queues, since the
	// constants of proportionality are small. Simultaneously, we want to
	// avoid a situation where a single worker ends up queuing O(heap)
	// work and then forever retains a queue of that size.
	const maxCap = 1 << 20 / goarch.PtrSize
	newCap := q.chain.head.cap * 2
	if newCap > maxCap {
		newCap = maxCap
	}
	newHead := newSpanSPMC11(newCap)
	if !q.tryDrain(newHead, n) {
		throw("failed to put span on newly-allocated spanSPMC")
	}
	q.chain.head.prev.StoreNoWB(unsafe.Pointer(newHead))
	q.chain.head = newHead
}

// tryDrain attempts to drain n spans from q's local queue to the chain.
//
// Returns whether it succeeded.
func (q *spanQueue11) tryDrain(r *spanSPMC11, n uint32) bool {
	if q.head+n > q.tail {
		throw("attempt to drain too many elements")
	}
	h := r.head.Load() // synchronize with consumers
	t := r.tail.Load()
	rn := t - h
	if rn+n <= r.cap {
		for i := uint32(0); i < n; i++ {
			*r.slot(t + i) = q.ring[(q.head+i)%uint32(len(q.ring))]
		}
		r.tail.Store(t + n) // Makes the items avail for consumption.
		q.head += n
		return true
	}
	return false
}

// tryGetFast attempts to get a span from the local queue, but may fail if it's empty,
// returning false.
func (q *spanQueue11) tryGetFast() objptr11 {
	if q.tail-q.head == 0 {
		return 0
	}
	s := q.ring[q.head%uint32(len(q.ring))]
	q.head++
	return s
}

// steal takes some spans from the ring chain of another span queue.
//
// q == q2 is OK.
func (q *spanQueue11) steal(q2 *spanQueue11) objptr11 {
	r := (*spanSPMC11)(q2.chain.tail.Load())
	if r == nil {
		return 0
	}
	for {
		// It's important that we load the next pointer
		// *before* popping the tail. In general, r may be
		// transiently empty, but if next is non-nil before
		// the pop and the pop fails, then r is permanently
		// empty, which is the only condition under which it's
		// safe to drop r from the chain.
		r2 := (*spanSPMC11)(r.prev.Load())

		// Try to refill from one of the rings
		if s := q.refill(r); s != 0 {
			return s
		}

		if r2 == nil {
			// This is the only ring. It's empty right
			// now, but could be pushed to in the future.
			return 0
		}

		// The tail of the chain has been drained, so move on
		// to the next ring. Try to drop it from the chain
		// so the next consumer doesn't have to look at the empty
		// ring again.
		if q2.chain.tail.CompareAndSwapNoWB(unsafe.Pointer(r), unsafe.Pointer(r2)) {
			r.dead.Store(true)
		}

		r = r2
	}
}

// refill takes some spans from r and puts them into q's local queue.
//
// One span is removed from the stolen spans and returned on success.
// Failure to steal returns a zero objptr.
//
// steal is thread-safe with respect to r.
func (q *spanQueue11) refill(r *spanSPMC11) objptr11 {
	if q.tail-q.head != 0 {
		throw("steal with local work available")
	}

	// Steal some spans.
	var n uint32
	for {
		h := r.head.Load() // load-acquire, synchronize with other consumers
		t := r.tail.Load() // load-acquire, synchronize with the producer
		n = t - h
		n = n - n/2
		if n == 0 {
			return 0
		}
		if n > r.cap { // read inconsistent h and t
			continue
		}
		n = min(n, uint32(len(q.ring)/2))
		for i := uint32(0); i < n; i++ {
			q.ring[i] = *r.slot(h + i)
		}
		if r.head.CompareAndSwap(h, h+n) {
			break
		}
	}

	// Update local queue head and tail to reflect new buffered values.
	q.head = 0
	q.tail = n

	// Pop off the head of the queue and return it.
	return q.tryGetFast()
}

// spanSPMC11 is a ring buffer of objptrs that represent spans.
// Accessed without a lock.
//
// Single-producer, multi-consumer. The only producer is the P that owns this
// queue, but any other P may consume from it.
//
// ## Invariants for memory management
//
// 1. All spanSPMCs are allocated from mheap_.spanSPMCAlloc.
// 2. All allocated spanSPMCs must be on the work.spanSPMCs list.
// 3. spanSPMCs may only be allocated if gcphase != _GCoff.
// 4. spanSPMCs may only be deallocated if gcphase == _GCoff.
//
// Invariants (3) and (4) ensure that we do not need to concern ourselves with
// tricky reuse issues that stem from not knowing when a thread is truly done
// with a spanSPMC11. For example, two threads could load the same spanSPMC11 from
// the tail of the chain. One thread is then paused while the other steals the
// last few elements off of it. It's not safe to free at that point since the
// other thread will still inspect that spanSPMC11, and we have no way of knowing
// without more complex and/or heavyweight synchronization.
//
// Instead, we rely on the global synchronization inherent to GC phases, and
// the fact that spanSPMCs are only ever used during the mark phase, to ensure
// memory safety. This means we temporarily waste some memory, but it's only
// until the end of the mark phase.
type spanSPMC11 struct {
	_ sys.NotInHeap

	// allnext is the link to the next spanSPMC on the work.spanSPMCs list.
	// This is used to find and free dead spanSPMCs. Protected by
	// work.spanSPMCs.lock.
	allnext *spanSPMC11

	// dead indicates whether the spanSPMC is no longer in use.
	// Protected by the CAS to the prev field of the spanSPMC pointing
	// to this spanSPMC. That is, whoever wins that CAS takes ownership
	// of marking this spanSPMC as dead. See spanQueue.steal for details.
	dead atomic.Bool

	// prev is the next link up a spanQueue's SPMC chain, from tail to head,
	// hence the name "prev." Set by a spanQueue's producer, cleared by a
	// CAS in spanQueue.steal.
	prev atomic.UnsafePointer // *spanSPMC

	// head, tail, cap, and ring together represent a fixed-size SPMC lock-free
	// ring buffer of size cap. The ring buffer contains objptr values.
	head atomic.Uint32
	tail atomic.Uint32
	cap  uint32 // cap(ring))
	ring *objptr11
}

// newSpanSPMC11 allocates and initializes a new spmc with the provided capacity.
//
// newSpanSPMC11 may override the capacity with a larger one if the provided one would
// waste memory.
func newSpanSPMC11(cap uint32) *spanSPMC11 {
	lock(&work.spanSPMCs11.lock)
	r := (*spanSPMC11)(mheap_.spanSPMCAlloc.alloc())
	r.allnext = work.spanSPMCs11.all
	work.spanSPMCs11.all = r
	unlock(&work.spanSPMCs11.lock)

	// If cap < the capacity of a single physical page, round up.
	pageCap := uint32(physPageSize / goarch.PtrSize) // capacity of a single page
	if cap < pageCap {
		cap = pageCap
	}
	if cap&(cap-1) != 0 {
		throw("spmc capacity must be a power of 2")
	}

	r.cap = cap
	ring := sysAlloc(uintptr(cap)*unsafe.Sizeof(objptr11(0)), &memstats.gcMiscSys, "GC span queue")
	atomic.StorepNoWB(unsafe.Pointer(&r.ring), ring)
	return r
}

// empty returns true if the spmc is empty.
//
// empty is thread-safe.
func (r *spanSPMC11) empty() bool {
	h := r.head.Load()
	t := r.tail.Load()
	return t == h
}

// deinit frees any resources the spanSPMC is holding onto and zeroes it.
func (r *spanSPMC11) deinit() {
	sysFree(unsafe.Pointer(r.ring), uintptr(r.cap)*unsafe.Sizeof(objptr11(0)), &memstats.gcMiscSys)
	r.ring = nil
	r.dead.Store(false)
	r.prev.StoreNoWB(nil)
	r.head.Store(0)
	r.tail.Store(0)
	r.cap = 0
}

// slot returns a pointer to slot i%r.cap.
func (r *spanSPMC11) slot(i uint32) *objptr11 {
	idx := uintptr(i & (r.cap - 1))
	return (*objptr11)(unsafe.Add(unsafe.Pointer(r.ring), idx*unsafe.Sizeof(objptr11(0))))
}

// freeDeadSpanSPMCs11 frees dead spanSPMCs back to the OS.
func freeDeadSpanSPMCs11() {
	// According to the SPMC memory management invariants, we can only free
	// spanSPMCs outside of the mark phase. We ensure we do this in two ways.
	//
	// 1. We take the work.spanSPMCs lock, which we need anyway. This ensures
	//    that we are non-preemptible. If this path becomes lock-free, we will
	//    need to become non-preemptible in some other way.
	// 2. Once we are non-preemptible, we check the gcphase, and back out if
	//    it's not safe.
	//
	// This way, we ensure that we don't start freeing if we're in the wrong
	// phase, and the phase can't change on us while we're freeing.
	//
	// TODO(go.dev/issue/75771): Due to the grow semantics in
	// spanQueue.drain, we expect a steady-state of around one spanSPMC per
	// P, with some spikes higher when Ps have more than one. For high
	// GOMAXPROCS, or if this list otherwise gets long, it would be nice to
	// have a way to batch work that allows preemption during processing.
	lock(&work.spanSPMCs11.lock)
	if gcphase != _GCoff || work.spanSPMCs11.all == nil {
		unlock(&work.spanSPMCs11.lock)
		return
	}
	rp := &work.spanSPMCs11.all
	for {
		r := *rp
		if r == nil {
			break
		}
		if r.dead.Load() {
			// It's dead. Deinitialize and free it.
			*rp = r.allnext
			r.deinit()
			mheap_.spanSPMCAlloc.free(unsafe.Pointer(r))
		} else {
			// Still alive, likely in some P's chain.
			// Skip it.
			rp = &r.allnext
		}
	}
	unlock(&work.spanSPMCs11.lock)
}

// tryStealSpan attempts to steal a span from another P's local queue.
//
// Returns a non-zero objptr on success.
func (w *gcWork) tryStealSpan11() objptr11 {
	pp := getg().m.p.ptr()

	for enum := stealOrder.start(cheaprand()); !enum.done(); enum.next() {
		if !work.spanqMask.read(enum.position()) {
			continue
		}
		p2 := allp[enum.position()]
		if pp == p2 {
			continue
		}
		if s := w.spanq11.steal(&p2.gcw.spanq11); s != 0 {
			return s
		}
		// N.B. This is intentionally racy. We may stomp on a mask set by
		// a P that just put a bunch of work into its local queue.
		//
		// This is OK because the ragged barrier in gcMarkDone will set
		// the bit on each P if there's local work we missed. This race
		// should generally be rare, since the window between noticing
		// an empty local queue and this bit being set is quite small.
		work.spanqMask.clear(int32(enum.position()))
	}
	return 0
}

// objptr11 consists of a span base and the index of the object in the span.
type objptr11 uintptr

func (p objptr11) spanBase() uintptr {
	return uintptr(p) &^ ((1 << gc.PageShift) - 1)
}

func (p objptr11) objIndex() uint16 {
	return uint16(p) & ((1 << gc.PageShift) - 1)
}
