// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// castagnoliUpdate updates the non-inverted crc with the given data.

// func castagnoliUpdate(crc uint32, p []byte) uint32
TEXT ·castagnoliUpdate(SB),NOSPLIT,$0-36
	MOVWU	crc+0(FP), R4		// a0 = CRC value
	MOVV	p+8(FP), R5		// a1 = data pointer
	MOVV	p_len+16(FP), R6	// a2 = len(p)

	SGT	$16, R6, R12
	BNE	R12, less_than_16
	AND	$15, R5, R12
	BEQ	R12, aligned

	// Process the first few bytes to 16-byte align the input.
	// t0 = 16 - t0. We need to process this many bytes to align.
	SUB	$1, R12
	XOR	$15, R12

	AND	$1, R12, R13
	BEQ	R13, align_2
	MOVB	(R5), R13
	CRCCWBW	R4, R13, R4
	ADDV	$1, R5
	ADDV	$-1, R6

align_2:
	AND	$2, R12, R13
	BEQ	R13, align_4
	MOVH	(R5), R13
	CRCCWHW	R4, R13, R4
	ADDV	$2, R5
	ADDV	$-2, R6

align_4:
	AND	$4, R12, R13
	BEQ	R13, aligned
	MOVW	(R5), R13
	CRCCWWW	R4, R13, R4
	ADDV	$4, R5
	ADDV	$-4, R6

align_8:
	AND	$8, R12, R13
	BEQ	R13, aligned
	MOVV	(R5), R13
	CRCCWVW	R4, R13, R4
	ADDV	$8, R5
	ADDV	$-8, R6

aligned:
	// The input is now 16-byte aligned and we can process 16-byte chunks.
	SGT	$16, R6, R12
	BNE	R12, less_than_16
	MOVV	(R5), R13
	MOVV	8(R5), R14
	CRCCWVW	R4, R13, R4
	CRCCWVW	R4, R14, R4
	ADDV	$16, R5
	ADDV	$-16, R6
	JMP	aligned

less_than_16:
	// We may have some bytes left over; process 8 bytes, then 4, then 2, then 1.
	AND	$8, R6, R12
	BEQ	R12, less_than_8
	MOVV	(R5), R13
	CRCCWVW	R4, R13, R4
	ADDV	$8, R5
	ADDV	$-8, R6

less_than_8:
	AND	$4, R6, R12
	BEQ	R12, less_than_4
	MOVW	(R5), R13
	CRCCWWW	R4, R13, R4
	ADDV	$4, R5
	ADDV	$-4, R6

less_than_4:
	AND	$2, R6, R12
	BEQ	R12, less_than_2
	MOVH	(R5), R13
	CRCCWHW	R4, R13, R4
	ADDV	$2, R5
	ADDV	$-2, R6

less_than_2:
	BEQ	R6, done
	MOVB	(R5), R13
	CRCCWBW	R4, R13, R4

done:
	MOVW	R4, ret+32(FP)
	RET

// ieeeUpdate updates the non-inverted crc with the given data.

// func ieeeUpdate(crc uint32, p []byte) uint32
TEXT ·ieeeUpdate(SB),NOSPLIT,$0-36
	MOVWU	crc+0(FP), R4		// a0 = CRC value
	MOVV	p+8(FP), R5		// a1 = data pointer
	MOVV	p_len+16(FP), R6	// a2 = len(p)

	SGT	$16, R6, R12
	BNE	R12, less_than_16
	AND	$15, R5, R12
	BEQ	R12, aligned

	// Process the first few bytes to 16-byte align the input.
	// t0 = 16 - t0. We need to process this many bytes to align.
	SUB	$1, R12
	XOR	$15, R12

	AND	$1, R12, R13
	BEQ	R13, align_2
	MOVB	(R5), R13
	CRCWBW	R4, R13, R4
	ADDV	$1, R5
	ADDV	$-1, R6

align_2:
	AND	$2, R12, R13
	BEQ	R13, align_4
	MOVH	(R5), R13
	CRCWHW	R4, R13, R4
	ADDV	$2, R5
	ADDV	$-2, R6

align_4:
	AND	$4, R12, R13
	BEQ	R13, aligned
	MOVW	(R5), R13
	CRCWWW	R4, R13, R4
	ADDV	$4, R5
	ADDV	$-4, R6

align_8:
	AND	$8, R12, R13
	BEQ	R13, aligned
	MOVV	(R5), R13
	CRCWVW	R4, R13, R4
	ADDV	$8, R5
	ADDV	$-8, R6

aligned:
	// The input is now 16-byte aligned and we can process 16-byte chunks.
	SGT	$16, R6, R12
	BNE	R12, less_than_16
	MOVV	(R5), R13
	MOVV	8(R5), R14
	CRCWVW	R4, R13, R4
	CRCWVW	R4, R14, R4
	ADDV	$16, R5
	ADDV	$-16, R6
	JMP	aligned

less_than_16:
	// We may have some bytes left over; process 8 bytes, then 4, then 2, then 1.
	AND	$8, R6, R12
	BEQ	R12, less_than_8
	MOVV	(R5), R13
	CRCWVW	R4, R13, R4
	ADDV	$8, R5
	ADDV	$-8, R6

less_than_8:
	AND	$4, R6, R12
	BEQ	R12, less_than_4
	MOVW	(R5), R13
	CRCWWW	R4, R13, R4
	ADDV	$4, R5
	ADDV	$-4, R6

less_than_4:
	AND	$2, R6, R12
	BEQ	R12, less_than_2
	MOVH	(R5), R13
	CRCWHW	R4, R13, R4
	ADDV	$2, R5
	ADDV	$-2, R6

less_than_2:
	BEQ	R6, done
	MOVB	(R5), R13
	CRCWBW	R4, R13, R4

done:
	MOVW	R4, ret+32(FP)
	RET
