// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file encapsulates some of the odd characteristics of the RISCV64
// instruction set, to minimize its interaction with the core of the
// assembler.

package arch

import (
	"cmd/internal/obj"
	"cmd/internal/obj/riscv"
)

// IsRISCV64AMO reports whether the op (as defined by a riscv.A*
// constant) is one of the AMO instructions that requires special
// handling.
func IsRISCV64AMO(op obj.As) bool {
	switch op {
	case riscv.ASCW, riscv.ASCD, riscv.AAMOSWAPW, riscv.AAMOSWAPD, riscv.AAMOADDW, riscv.AAMOADDD,
		riscv.AAMOANDW, riscv.AAMOANDD, riscv.AAMOORW, riscv.AAMOORD, riscv.AAMOXORW, riscv.AAMOXORD,
		riscv.AAMOMINW, riscv.AAMOMIND, riscv.AAMOMINUW, riscv.AAMOMINUD,
		riscv.AAMOMAXW, riscv.AAMOMAXD, riscv.AAMOMAXUW, riscv.AAMOMAXUD:
		return true
	}
	return false
}

// IsRISCV64XThead reports whether the op (as defined by a riscv.A*
// constant) is one of the XThead extensions instructions that requires
// special handling.
func IsRISCV64XThead(op obj.As) bool {
	switch op {
	case
		// 3. XTheadCmo provides instructions for cache management.
		riscv.ATHDCACHECALL, riscv.ATHDCACHECIALL, riscv.ATHDCACHEIALL, riscv.ATHDCACHECPA,
		riscv.ATHDCACHECIPA, riscv.ATHDCACHEIPA, riscv.ATHDCACHECVA, riscv.ATHDCACHECIVA,
		riscv.ATHDCACHEIVA, riscv.ATHDCACHECSW, riscv.ATHDCACHECISW, riscv.ATHDCACHEISW,
		riscv.ATHDCACHECPAL1, riscv.ATHDCACHECVAL1, riscv.ATHICACHEIALL, riscv.ATHICACHEIALLS,
		riscv.ATHICACHEIPA, riscv.ATHICACHEIVA, riscv.ATHL2CACHECALL, riscv.ATHL2CACHECIALL,
		riscv.ATHL2CACHEIALL, riscv.ATHSFENCEVMAS,
		// 4. XTheadSync provides instructions for multi-processor synchronization.
		riscv.ATHSYNC, riscv.ATHSYNCS, riscv.ATHSYNCI, riscv.ATHSYNCIS:
		return true
	}
	return false
}
