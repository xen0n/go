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

	// Moves.
	// Memory loads.
	MOV	(R4), R12				// 8c00c028
	MOV	233(R4), R12				// 8ca4c328
	MOVB	(R4), R12				// 8c000028
	MOVB	233(R4), R12				// 8ca40328
	MOVBU	(R4), R12				// 8c00002a
	MOVBU	233(R4), R12				// 8ca4032a
	MOVH	(R4), R12				// 8c004028
	MOVH	233(R4), R12				// 8ca44328
	MOVHU	(R4), R12				// 8c00402a
	MOVHU	233(R4), R12				// 8ca4432a
	MOVW	(R4), R12				// 8c008028
	MOVW	233(R4), R12				// 8ca48328
	MOVWU	(R4), R12				// 8c00802a
	MOVWU	233(R4), R12				// 8ca4832a
	FMOVS	(R4), F1				// 8100002b
	FMOVS	233(R4), F1				// 81a4032b
	FMOVD	(R4), F1				// 8100802b
	FMOVD	233(R4), F1				// 81a4832b

	// Memory stores.
	MOV	R12, (R4)				// 8c00c029
	MOV	R12, 233(R4)				// 8ca4c329
	MOVB	R12, (R4)				// 8c000029
	MOVB	R12, 233(R4)				// 8ca40329
	MOVH	R12, (R4)				// 8c004029
	MOVH	R12, 233(R4)				// 8ca44329
	MOVW	R12, (R4)				// 8c008029
	MOVW	R12, 233(R4)				// 8ca48329
	FMOVS	F1, (R4)				// 8100402b
	FMOVS	F1, 233(R4)				// 81a4432b
	FMOVD	F1, (R4)				// 8100c02b
	FMOVD	F1, 233(R4)				// 81a4c32b

	// Register-to-register moves.
	MOV	R9, R10					// 2a011500
	MOVB	R11, R12				// 6c5d0000
	MOVBU	R13, R14				// aefd4303
	MOVH	R15, R16				// f0590000
	MOVHU	R17, R18				// 3202cf00
	MOVW	R19, R20				// 74028002
	MOVWU	R4, R5					// 8500df00

	// These two are actually real instructions, but include them
	// regardless.
	FMOVS	F6, F7					// c7941401
	FMOVD	F8, F9					// 09991401
