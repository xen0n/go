// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong

import (
	"cmd/internal/obj"
	"fmt"
)

// addrToReg extracts the register from an Addr, handling special Addr.Names.
func addrToReg(a obj.Addr) int16 {
	switch a.Name {
	case obj.NAME_PARAM, obj.NAME_AUTO:
		return REG_SP
	}
	return a.Reg
}

// movToLoad converts a MOV mnemonic into the corresponding load instruction.
func movToLoad(as obj.As) obj.As {
	switch as {
	case AMOV:
		return ALDD
	case AMOVB:
		return ALDB
	case AMOVH:
		return ALDH
	case AMOVW:
		return ALDW
	case AMOVBU:
		return ALDBU
	case AMOVHU:
		return ALDHU
	case AMOVWU:
		return ALDWU

	// It is syntactic sugar to allow FMOVs to act as loads/stores.
	case AFMOVS:
		return AFLDS
	case AFMOVD:
		return AFLDD

	default:
		panic(fmt.Sprintf("%v is not a MOV", as))
	}
}

// movToStore converts a MOV mnemonic into the corresponding store instruction.
func movToStore(as obj.As) obj.As {
	switch as {
	case AMOV:
		return ASTD
	case AMOVB:
		return ASTB
	case AMOVH:
		return ASTH
	case AMOVW:
		return ASTW
	case AMOVBU, AMOVHU, AMOVWU:
		panic(fmt.Sprintf("%v cannot be used to represent stores", as))

	// It is syntactic sugar to allow FMOVs to act as loads/stores.
	case AFMOVS:
		return AFSTS
	case AFMOVD:
		return AFSTD

	default:
		panic(fmt.Sprintf("%v is not a MOV", as))
	}
}

// rewriteMOV is called during preprocess to rewrite the given MOV
// pseudo-instruction with concrete instruction(s).
func rewriteMOV(ctxt *obj.Link, newprog obj.ProgAlloc, p *obj.Prog) {
	switch p.As {
	case AMOV, AMOVB, AMOVBU, AMOVH, AMOVHU, AMOVW, AMOVWU, AFMOVS, AFMOVD:
	default:
		panic(fmt.Sprintf("%+v is not a MOV pseudo-instruction", p.As))
	}

	switch p.From.Type {
	case obj.TYPE_MEM: // MOV c(Rs), Rd -> L $c, Rs, Rd
		switch p.From.Name {
		case obj.NAME_AUTO, obj.NAME_PARAM, obj.NAME_NONE:
			if p.To.Type != obj.TYPE_REG {
				ctxt.Diag("unsupported load at %v", p)
			}

			offset := p.From.Offset
			reg := addrToReg(p.From)

			p.As = movToLoad(p.As)
			p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: offset}
			p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: reg}})

		case obj.NAME_EXTERN, obj.NAME_STATIC:
			// TODO: Implement.
			ctxt.Diag("TODO")

		default:
			ctxt.Diag("unsupported name %d for %v", p.From.Name, p)
		}

	case obj.TYPE_REG:
		switch p.To.Type {
		case obj.TYPE_REG:
			rewriteRegRegMOV(newprog, p)

		case obj.TYPE_MEM: // MOV Rs, c(Rd) -> S $c, Rd, Rs
			switch p.As {
			case AMOVBU, AMOVHU, AMOVWU:
				ctxt.Diag("unsupported unsigned store at %v", p)
			}
			switch p.To.Name {
			case obj.NAME_AUTO, obj.NAME_PARAM, obj.NAME_NONE:
				offset := p.To.Offset
				reg := addrToReg(p.To)

				p.As = movToStore(p.As)
				p.To = p.From
				p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: offset}
				p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: reg}})

			case obj.NAME_EXTERN, obj.NAME_STATIC:
				// TODO: Implement.
				ctxt.Diag("TODO")

			default:
				ctxt.Diag("unsupported name %d for %v", p.From.Name, p)
			}

		default:
			ctxt.Diag("unsupported MOV at %v", p)
		}

	case obj.TYPE_CONST:
		if p.As != AMOV {
			ctxt.Diag("%v: constant load must use MOV", p)
		}
		if p.To.Type != obj.TYPE_REG {
			ctxt.Diag("%v: constant load must target register", p)
		}

		materializeSImm(newprog, p, p.From.Offset, p.To.Reg, true)

	case obj.TYPE_ADDR: // MOV $sym+off(SP/SB), R
		// TODO: Implement.
		ctxt.Diag("TODO")

	default:
		ctxt.Diag("unsupported MOV at %v", p)
	}
}

func rewriteRegRegMOV(newprog obj.ProgAlloc, p *obj.Prog) {
	switch p.As {
	case AFMOVS, AFMOVD:
		// These are already native instructions, nothing left to do.

	case AMOV: // MOV Ra, Rb -> OR ZERO, Ra, Rb
		p.As = AOR
		p.SetRestArgs([]obj.Addr{p.From})
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_ZERO}

	case AMOVB: // MOVB Ra, Rb -> SEXTB Ra, Rb
		p.As = ASEXTB

	case AMOVBU: // MOVBU Ra, Rb -> ANDI $255, Ra, Rb
		p.As = AANDI
		p.SetRestArgs([]obj.Addr{p.From})
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 0xff}

	case AMOVH: // MOVH Ra, Rb -> SEXTH Ra, Rb
		p.As = ASEXTH

	case AMOVW: // MOVW Ra, Rb -> ADDIW $0, Ra, Rb
		p.As = AADDIW
		p.SetRestArgs([]obj.Addr{p.From})
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 0}

	case AMOVHU, AMOVWU: // MOV[HW]U Ra, Rb -> BSTRPICKD $bitwidth-1, $0, Ra, Rb
		var bitwidth int64
		if p.As == AMOVHU {
			bitwidth = 15
		} else {
			bitwidth = 31
		}

		p.As = ABSTRPICKD
		p.SetRestArgs([]obj.Addr{
			{Type: obj.TYPE_CONST, Offset: 0},
			p.From,
		})
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: bitwidth}
	}
}
