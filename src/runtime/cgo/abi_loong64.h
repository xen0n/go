// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Macros for transitioning from the host ABI to Go ABI0.
//
// These save the frame pointer, so in general, functions that use
// these should have zero frame size to suppress the automatic frame
// pointer, though it's harmless to not do this.

#define PUSH_REGS_FLOAT_PART() \
	MOVD	F24, 120(R3) \
	MOVD	F25, 128(R3) \
	MOVD	F26, 136(R3) \
	MOVD	F27, 144(R3) \
	MOVD	F28, 152(R3) \
	MOVD	F29, 160(R3) \
	MOVD	F30, 168(R3) \
	MOVD	F31, 176(R3)

#define POP_REGS_FLOAT_PART() \
	MOVD	120(R3), F24 \
	MOVD	128(R3), F25 \
	MOVD	136(R3), F26 \
	MOVD	144(R3), F27 \
	MOVD	152(R3), F28 \
	MOVD	160(R3), F29 \
	MOVD	168(R3), F30 \
	MOVD	176(R3), F31

#define PUSH_REGS_HOST_TO_ABI0() \
	MOVV	R23, 24(R3) \
	MOVV	R24, 32(R3) \
	MOVV	R25, 40(R3) \
	MOVV	R26, 48(R3) \
	MOVV	R27, 56(R3) \
	MOVV	R28, 64(R3) \
	MOVV	R29, 72(R3) \
	MOVV	R30, 80(R3) \
	MOVV	R31, 88(R3) \
	MOVV	R3, 96(R3) \
	MOVV	g, 104(R3) \
	MOVV	R1, 112(R3) \
	PUSH_REGS_FLOAT_PART()

#define POP_REGS_HOST_TO_ABI0() \
	MOVV	24(R3), R23 \
	MOVV	32(R3), R24 \
	MOVV	40(R3), R25 \
	MOVV	48(R3), R26 \
	MOVV	56(R3), R27 \
	MOVV	64(R3), R28 \
	MOVV	72(R3), R29 \
	MOVV	80(R3), R30 \
	MOVV	88(R3), R31 \
	MOVV	96(R3), R3 \
	MOVV	104(R3), g \
	MOVV	112(R3), R1 \
	POP_REGS_FLOAT_PART()
