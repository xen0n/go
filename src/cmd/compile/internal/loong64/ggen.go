// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong64

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/objw"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/obj/loong"
)

func zeroRange(pp *objw.Progs, p *obj.Prog, off, cnt int64, _ *uint32) *obj.Prog {
	if cnt == 0 {
		return p
	}

	// Adjust the frame to account for LR.
	off += base.Ctxt.FixedFrameSize()

	if cnt < int64(4*types.PtrSize) {
		for i := int64(0); i < cnt; i += int64(types.PtrSize) {
			// STD	$(off + i), SP, ZERO
			p = pp.Append(p, loong.ASTD, obj.TYPE_CONST, 0, off+i, obj.TYPE_REG, loong.REG_ZERO, 0)
			p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: loong.REG_SP}})
		}
		return p
	}

	// TODO: Add a duff zero implementation for medium sized ranges.

	// Loop, zeroing pointer width bytes at a time.
	// ADDD	$(off), SP, T0
	// ADDD	$(cnt), T0, T1
	// loop:
	// 	STD	$0, T0, ZERO
	// 	ADDID	$Widthptr, ZERO, T0
	//	BNE	T0, T1, loop
	p = pp.Append(p, loong.AADDD, obj.TYPE_CONST, 0, off, obj.TYPE_REG, loong.REG_T0, 0)
	p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: loong.REG_SP}})
	p = pp.Append(p, loong.AADDD, obj.TYPE_CONST, 0, cnt, obj.TYPE_REG, loong.REG_T1, 0)
	p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: loong.REG_T0}})
	p = pp.Append(p, loong.ASTD, obj.TYPE_CONST, 0, 0, obj.TYPE_REG, loong.REG_ZERO, 0)
	p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: loong.REG_T0}})
	loop := p
	p = pp.Append(p, loong.AADDID, obj.TYPE_CONST, 0, int64(types.PtrSize), obj.TYPE_REG, loong.REG_T0, 0)
	p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: loong.REG_ZERO}})
	p = pp.Append(p, loong.ABNE, obj.TYPE_REG, loong.REG_T0, 0, obj.TYPE_BRANCH, 0, 0)
	p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: loong.REG_T1}})
	p.To.SetTarget(loop)
	return p
}
