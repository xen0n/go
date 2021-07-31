// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong

import (
	"cmd/internal/obj"
	"fmt"
)

// All unary instructions which write to their arguments (as opposed to reading
// from them) go here. The assembly parser uses this information to populate
// its AST in a semantically reasonable way.
//
// Any instructions not listed here are assumed to either be non-unary or to read
// from its argument.
var unaryDst = map[obj.As]bool{
	// No supported LoongArch instruction falls inside this.
}

type instruction struct {
	as   obj.As // Assembler opcode
	rd   uint32 // Destination register
	rj   uint32 // Source register 1
	rk   uint32 // Source register 2
	ra   uint32 // Source register 3
	imm1 int64  // Immediate 1
	imm2 int64  // Immediate 2
}

func isDirectiveInsn(as obj.As) bool {
	switch as {
	// This list is taken from riscv/obj.go, these are the directives we
	// could encounter while assembling.
	case obj.AFUNCDATA, obj.APCDATA, obj.ATEXT, obj.ANOP, obj.ADUFFZERO, obj.ADUFFCOPY:
		return true
	}
	return false
}

func isPseudoInsn(as obj.As) bool {
	switch as {
	case AWORD, AMOV, AMOVB, AMOVBU, AMOVH, AMOVHU, AMOVW, AMOVWU:
		return true
	}
	return false
}

func (insn *instruction) length() int {
	// Directives have no encoding, hence no length.
	if isDirectiveInsn(insn.as) {
		return 0
	}

	// LoongArch instructions will remain fixed-length in the foreseeable future.
	return 4
}

func (insn *instruction) validate(ctxt *obj.Link) {
	// Directives need no encoding validation.
	if isDirectiveInsn(insn.as) {
		return
	}

	// Special-case pseudo-instructions.
	if isPseudoInsn(insn.as) {
		switch insn.as {
		case AWORD:
			// Treat the raw value specially as a 32-bit unsigned integer.
			// Nobody wants to enter negative machine code.
			if insn.imm1 < 0 || 1<<32 <= insn.imm1 {
				ctxt.Diag(
					"%v\timmediate in raw position cannot be larger than 32 bits but got %d",
					insn.as,
					insn.imm1,
				)
			}
			return

		default:
			// Other pseudo-instructions should not exist at this point.
			ctxt.Diag("%v\tnon-rewritten pseudo-instruction", insn.as)
			return
		}
	}

	enc, err := encodingForAs(insn.as)
	if err != nil {
		ctxt.Diag(err.Error())
		return
	}

	if enc.fmt < 1 || int(enc.fmt) >= len(validators) {
		ctxt.Diag(fmt.Sprintf("unknown insn format %d", enc.fmt))
		return
	}

	err = validators[enc.fmt](insn)
	if err != nil {
		ctxt.Diag(err.Error())
	}
}

func (insn *instruction) encode() (uint32, error) {
	// Special-case the WORD pseudo-instruction; all other instruction
	// should be concrete at this point.
	switch insn.as {
	case AWORD:
		return uint32(insn.imm1), nil
	}

	return insn.encodeReal()
}

// encodingForAs returns the encoding for an obj.As.
func encodingForAs(as obj.As) (*encoding, error) {
	if base := as &^ obj.AMask; base != obj.ABaseLoong && base != 0 {
		return nil, fmt.Errorf("encodingForAs: not a LoongArch instruction %s", as)
	}
	asi := as & obj.AMask
	if int(asi) >= len(encodings) {
		return nil, fmt.Errorf("encodingForAs: bad LoongArch instruction %s", as)
	}
	return &encodings[asi], nil
}

func regVal(r, min, max uint32) uint32 {
	if r < min || r > max {
		panic(fmt.Sprintf("register out of range, want %d < %d < %d", min, r, max))
	}
	return r - min
}

// regInt returns an integer register.
func regInt(r uint32) uint32 {
	return regVal(r, REG_R0, REG_R31)
}

// regFP returns a float register.
func regFP(r uint32) uint32 {
	return regVal(r, REG_F0, REG_F31)
}

// regFCC returns a floating-point condition code register.
func regFCC(r uint32) uint32 {
	return regVal(r, REG_FCC0, REG_FCC7)
}

// simmFits reports whether immediate value x fits in nbits bits
// as a signed integer.
func simmFits(x int64, nbits uint) bool {
	nbits--
	var min int64 = -1 << nbits
	var max int64 = 1<<nbits - 1
	return min <= x && x <= max
}

// uimmFits reports whether immediate value x fits in nbits bits
// as an unsigned integer.
func uimmFits(x int64, nbits uint) bool {
	var max int64 = 1<<nbits - 1
	return 0 <= x && x <= max
}

func wantSignedImm(as obj.As, imm int64, nbits uint) error {
	if !simmFits(imm, nbits) {
		return fmt.Errorf("%v\tsigned immediate cannot be larger than %d bits but got %d", as, nbits, imm)
	}

	return nil
}

func wantUnsignedImm(as obj.As, imm int64, nbits uint) error {
	if !uimmFits(imm, nbits) {
		return fmt.Errorf("%v\tunsigned immediate cannot be larger than %d bits but got %d", as, nbits, imm)
	}

	return nil
}

func wantReg(as obj.As, descr string, r, min, max uint32) error {
	if r < min || r > max {
		var suffix string
		if r != obj.REG_NONE {
			suffix = fmt.Sprintf(" but got non-%s register %s", descr, RegName(int(r)))
		}
		return fmt.Errorf("%v\texpected %s register%s", as, descr, suffix)
	}

	return nil
}

// wantIntReg checks that r is an integer register.
func wantIntReg(as obj.As, r uint32) error {
	return wantReg(as, "integer", r, REG_R0, REG_R31)
}

// wantFPReg checks that r is a floating-point register.
func wantFPReg(as obj.As, r uint32) error {
	return wantReg(as, "float", r, REG_F0, REG_F31)
}

// wantFCCReg checks that r is a floating-point condition code register.
func wantFCCReg(as obj.As, r uint32) error {
	return wantReg(as, "fcc", r, REG_FCC0, REG_FCC7)
}

// instructionsForProg returns the machine instructions for an *obj.Prog.
func instructionsForProg(p *obj.Prog) []*instruction {
	if isDirectiveInsn(p.As) {
		return nil
	}

	// Special-case the WORD pseudo-instruction.
	if p.As == AWORD {
		return []*instruction{
			{as: p.As, imm1: p.From.Offset},
		}
	}

	enc, err := encodingForAs(p.As)
	if err != nil {
		// Let the error through for it to manifest during assembly.
		return []*instruction{
			{as: p.As},
		}
	}

	insn := instruction{
		as: p.As,
	}

	// Handle problematic instructions.
	switch p.As {
	case AASRTLE, AASRTGT:
		insn.rk = uint32(p.From.Reg)
		insn.rj = uint32(p.RestArgs[0].Reg)

	case ALDPTE:
		insn.imm1 = p.From.Offset
		insn.rj = uint32(p.RestArgs[0].Reg)

	case ARDTICKLW, ARDTICKHW, ARDTICKD:
		insn.rj = uint32(p.To.Reg)
		insn.rd = uint32(p.GetTo2().Reg)

	case AFCSRWR:
		insn.rj = uint32(p.To.Reg)
		insn.imm1 = p.From.Offset

	case ACACOP, APRELD:
		insn.imm2 = p.From.Offset
		insn.imm1 = p.RestArgs[0].Offset
		insn.rj = uint32(p.RestArgs[1].Reg)

	case APRELDX, ATLBINV:
		insn.imm1 = p.From.Offset
		insn.rk = uint32(p.RestArgs[0].Reg)
		insn.rj = uint32(p.RestArgs[1].Reg)

	case AALSLW, AALSLWU, AALSLD, ABYTEPICKW, ABYTEPICKD:
		insn.imm1 = p.From.Offset
		insn.rk = uint32(p.RestArgs[0].Reg)
		insn.rj = uint32(p.RestArgs[1].Reg)
		insn.rd = uint32(p.To.Reg)

	default:
		// Encoding is reasonably regular at this point.
		switch enc.fmt.arity() {
		case 0:
			// Nothing left to do.

		case 1:
			insn.rd = uint32(p.To.Reg)
			insn.imm1 = p.From.Offset

		case 2:
			insn.rj = uint32(p.From.Reg)
			insn.rd = uint32(p.To.Reg)
			insn.imm1 = p.From.Offset

		case 3:
			insn.rk = uint32(p.From.Reg)
			insn.rj = uint32(p.RestArgs[0].Reg)
			insn.rd = uint32(p.To.Reg)
			insn.imm1 = p.From.Offset

		case 4:
			insn.ra = uint32(p.From.Reg)
			insn.rk = uint32(p.RestArgs[0].Reg)
			insn.rj = uint32(p.RestArgs[1].Reg)
			insn.rd = uint32(p.To.Reg)
			insn.imm2 = p.From.Offset
			insn.imm1 = p.RestArgs[0].Offset
		}
	}

	return []*instruction{&insn}
}
