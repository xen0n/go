// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package loong implements a LoongArch assembler, supporting instructions from the LoongArch Manual
Volume 1, version 1.00.

This package takes the instruction definitions from the community-maintained repository at
https://github.com/loongson-community/loongarch-opcodes; see its README for more details.

The assembly syntax used by Go is different from that in the official manuals, but we can still
follow the general rules to map between them. This document assumes some knowledge of LoongArch
assembler, and will focus on the differences; please refer to the official manuals for details not
covered by this document.

In the examples below, the Go assembly is on the left, the same operation rendered in official
manual syntax on the right.


Instructions mnemonics mapping rules

1. For nearly every instruction, remove all "." and "_" characters from official mnemonics for
the correct Go mnemonics.

  Examples:
    CLOW      R12, R13         <=>     clo.w       r13, r12
    MULWDWU   R4, R5, R6       <=>     mulw.d.wu   r6, r5, r4
    AMSWAPDBD R7, R8, R9       <=>     amswap_db.d r9, r8, r7

2. Some of the mnemonics have been tweaked to fix inconsistencies observed in the current version
of official manual. These are documented in the README document of the loongarch-opcodes project.

  Examples:
    BGT     R4, R5, label      <=>     blt        r4, r5, label
    FCSRRD  $0, R6             <=>     movfcsr2gr r6, r0
    RDTICKD R7, R8             <=>     rdtime.d   r8, r7

3. Go uses the MOV family of pseudo-instructions for memory accesses and constant loads.
Constant loads can only use the native-width form, that is MOV.

  Examples:
    MOV   (R4), R5             <=>     ld.d  r5, r4, 0
    MOV   4(R4), R5            <=>     ld.d  r5, r4, 4
    MOVB  (R4), R5             <=>     ld.b  r5, r4, 0
    MOVBU (R4), R5             <=>     ld.bu r5, r4, 0
    MOVH  (R4), R5             <=>     ld.h  r5, r4, 0
    MOVHU (R4), R5             <=>     ld.hu r5, r4, 0
    MOVW  (R4), R5             <=>     ld.w  r5, r4, 0
    MOVWU (R4), R5             <=>     ld.wu r5, r4, 0
    MOV   R4, (R5)             <=>     st.d  r4, r5, 0
    MOVB  R4, (R5)             <=>     st.b  r4, r5, 0
    MOVH  R4, (R5)             <=>     st.h  r4, r5, 0
    MOVW  R4, (R5)             <=>     st.w  r4, r5, 0


Register mapping rules

1. All integer register names are written as Rn, floating-point ones Fn, and
floating-point condition code ones FCCn.

2. The R21 register is reserved in the LA64 ABI, so you cannot use it in code at all;
neither "R21" nor "X" (the dubious name given by early WIP binutils port) will work.

Similarly, the g register is reserved in the Go ABI too, so "R31" and "S8" do not work either.
You can refer to it using the name "g", though.

3. ABI names of registers generally just work; with upper-case letters, of course.

You can write things like "A0", "T1", "FS2" instead of their raw names for readability.
There is, however, some exceptions to pay attention to:

(1) R22 is the FP register in LA64 ABI, but the "FP" name is reserved in Go assembler to mean
the virtual FP register. Just use "R22" to refer to that register.

(2) The Go ABI uses S7 for CTXT, and T8 for TMP. These two names can be used too.

(3) The early LoongArch toolchain ports from Loongson implement V0/V1 as aliases to A0/A1, possibly
in an attempt to be compatible with some unknown MIPS legacy. This is not done here.


Order of operands

In Go assembler, operands of most instructions are written in reverse order, that is, destination
operand comes last.

  Examples:
    CLOW   R4, R5              <=>     clo.w   r5, r4
    ADDD   R6, R7, R8          <=>     add.d   r8, r7, r6
    FMADDS F3, F2, F1, F0      <=>     fmadd.s f0, f1, f2, f3

Special Cases.

(1) Jumps are written with the destination (label) coming last, as this is the Go convention.

  Examples:
    BNE     R4, R5, label      <=>     bne        r4, r5, label

(2) Note that the "normal" order of operands is the one defined in the loongarch-opcodes project:
all registers before all immediates, from LSB to MSB direction in each group. This produces
different order for some instructions; note the imperfect "reversal" in the following examples.

  Examples:
    BSTRPICKD $15, $8, R9, R10 <=>     bstrpick.d r10, r9, 15, 8
    PRELD     $1, $2, R11      <=>     preld      2, r11, 1


Operands

There is no "processing" of immediates whatsoever, unlike in the Loongson binutils port.

  Examples:
    ALSLD  $3, R4, R5, R6      <=>     alsl.d  r6, r5, r4, 4
    LDOX4D $-100, R7, R8       <=>     ldptr.d r8, r7, -400


Syntactic sugar

1. If the second-to-last operand is the same as the last operand, one of them can be elided
for brevity.

  Examples:
    CLOW  R4                   <=>     clo.w  r4, r4
    ADDID $233, R5             <=>     addi.d r5, r5, 233
    ALSLD $3, R6, R7           <=>     alsl.d r7, r7, r6, 4

*/
package loong
