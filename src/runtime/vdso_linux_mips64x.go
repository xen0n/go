// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux
// +build mips64 mips64le

package runtime

const (
	// vdsoArrayMax is the byte-size of a maximally sized array on this architecture.
	// See cmd/compile/internal/mips64/galign.go arch.MAXWIDTH initialization.
	vdsoArrayMax = 1<<50 - 1
)

// key and version at man 7 vdso : mips
var vdsoLinuxVersion = vdsoVersionKey{"LINUX_2.6", 0x3ae75f6}

var vdsoSymbolKeys = []vdsoSymbolKey{
	{"__vdso_clock_gettime", 0xd35ec75, 0x6e43a318, &vdsoClockgettimeSym},
}

// initialize to fall back to syscall
var (
	vdsoClockgettimeSym uintptr = 0
)
