// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong

import (
	"cmd/internal/obj"
)

const (
	totalLowWidth    = 12
	totalUpperWidth  = 32
	totalHigherWidth = 52
)

// materializeSImm materializes the signed immediate imm in register reg,
// inserting the generated code sequence after p or replacing p altogether
// depending on rewriteP.
func materializeSImm(newprog obj.ProgAlloc, p *obj.Prog, imm int64, reg int16, overwriteP bool) {
	if !overwriteP {
		p = obj.Appendp(p, newprog)
	}

	if simmFits(imm, totalLowWidth) {
		// ADDIW $imm, ZERO, reg
		p.As = AADDIW
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = imm
		if overwriteP {
			p.RestArgs = nil
		}
		p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: REG_ZERO}})
		p.To.Type = obj.TYPE_REG
		p.To.Reg = reg
		return
	}

	chopped := chopSImm(imm)

	// common sequence:
	//
	// LU12IW $upper, reg
	// ORI $low, reg, reg  if low != 0

	p.As = ALU12IW
	p.From.Type = obj.TYPE_CONST
	p.From.Offset = int64(chopped.upper)
	p.To.Type = obj.TYPE_REG
	p.To.Reg = reg

	if chopped.low != 0 {
		p = obj.Appendp(p, newprog)
		p.As = AORI
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = int64(chopped.low)
		p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: reg}})
		p.To.Type = obj.TYPE_REG
		p.To.Reg = reg
	}

	if simmFits(imm, totalUpperWidth) {
		return
	}

	// LU32ID $higher, reg
	p = obj.Appendp(p, newprog)
	p.As = ALU32ID
	p.From.Type = obj.TYPE_CONST
	p.From.Offset = int64(chopped.higher)
	p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: reg}})
	p.To.Type = obj.TYPE_REG
	p.To.Reg = reg

	if simmFits(imm, totalHigherWidth) {
		return
	}

	// LU52ID $top, reg, reg
	p = obj.Appendp(p, newprog)
	p.As = ALU52ID
	p.From.Type = obj.TYPE_CONST
	p.From.Offset = int64(chopped.top)
	p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: reg}})
	p.To.Type = obj.TYPE_REG
	p.To.Reg = reg
}

type choppedSImm struct {
	// 11..0 bits of target immediate, to be loaded with ORI.
	low uint32
	// 31..12 bits of target immediate, to be loaded with LU12IW.
	upper int32
	// 51..32 bits of target immediate, to be loaded with LU32ID.
	higher int32
	// 63..52 bits of target immediate, to be loaded with LU52ID.
	top int32
}

func chopSImm(imm int64) choppedSImm {
	low := uint32(imm) & 0xfff
	upper := int32(adjustSImmForSlot(imm>>totalLowWidth&0xfffff, 20))
	higher := int32(adjustSImmForSlot(imm>>totalUpperWidth&0xfffff, 20))
	top := int32(adjustSImmForSlot(imm>>totalHigherWidth, 12))

	return choppedSImm{
		low:    low,
		upper:  upper,
		higher: higher,
		top:    top,
	}
}

func adjustSImmForSlot(x int64, width int) int64 {
	maxPositiveVal := int64(1)<<(width-1) - 1
	if x > maxPositiveVal {
		return -(1<<width - x)
	}
	return x
}
