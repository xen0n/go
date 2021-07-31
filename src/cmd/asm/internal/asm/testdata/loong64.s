// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "../../../../../runtime/textflag.h"

TEXT asmtest(SB),DUPOK|NOSPLIT,$0
start:

	JIRL	$0, RA, ZERO	// JIRL $0, R1, R0	// 2000004c

	// Arbitrary bytes (entered in little-endian mode)
	WORD	$0x12345678	// WORD $305419896	// 78563412
	WORD	$0x9abcdef0	// WORD $2596069104	// f0debc9a

	// Syntactic sugar: insns ending with two identical operands can have
	// one operand elided.
	CLOD	R4					// 84200000
	ADDD	R5, R6					// c6941000
	ALSLD	$3, R7, R8				// 089d2d00
