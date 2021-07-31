// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong

import (
	"cmd/internal/obj"
	"cmd/internal/sys"
)

var Linkloong64 = obj.LinkArch{
	Arch:           sys.ArchLoong64,
	Init:           instinit,
	Preprocess:     preprocess,
	Assemble:       assemble,
	Progedit:       progedit,
	UnaryDst:       unaryDst,
	DWARFRegisters: loong64DWARFRegisters,
}

func instinit(ctxt *obj.Link) {}

// progedit is called individually for each *obj.Prog. It normalizes instruction
// formats and eliminates as many pseudo-instructions as possible.
func progedit(ctxt *obj.Link, p *obj.Prog, newprog obj.ProgAlloc) {
	canonicalizeInsnArityForProg(p)

	// TODO: Implement.
}

// setPCs sets the Pc field in all instructions reachable from p.
// It uses pc as the initial value.
func setPCs(p *obj.Prog, pc int64) {
	for ; p != nil; p = p.Link {
		p.Pc = pc
		for _, insn := range instructionsForProg(p) {
			pc += int64(insn.length())
		}
	}
}

// preprocess generates prologue and epilogue code, computes PC-relative branch
// and jump offsets, and resolves pseudo-registers.
//
// preprocess is called once per linker symbol.
//
// When preprocess finishes, all instructions in the symbol are either
// concrete, real LoongArch instructions or directive pseudo-ops like TEXT,
// PCDATA, and FUNCDATA.
func preprocess(ctxt *obj.Link, cursym *obj.LSym, newprog obj.ProgAlloc) {
	if cursym.Func().Text == nil || cursym.Func().Text.Link == nil {
		return
	}

	// Generate the prologue.
	text := cursym.Func().Text
	if text.As != obj.ATEXT {
		ctxt.Diag("preprocess: found symbol that does not start with TEXT directive")
		return
	}

	stacksize := text.To.Offset
	if stacksize == -8 {
		// Historical way to mark NOFRAME.
		text.From.Sym.Set(obj.AttrNoFrame, true)
		stacksize = 0
	}
	if stacksize < 0 {
		ctxt.Diag("negative frame size %d - did you mean NOFRAME?", stacksize)
	}
	if text.From.Sym.NoFrame() {
		if stacksize != 0 {
			ctxt.Diag("NOFRAME functions must have a frame size of 0, not %d", stacksize)
		}
	}

	cursym.Func().Args = text.To.Val.(int32)
	cursym.Func().Locals = int32(stacksize)

	// TODO: Implement.

	setPCs(cursym.Func().Text, 0)

	// Validate all instructions - this provides nice error messages.
	for p := cursym.Func().Text; p != nil; p = p.Link {
		for _, ins := range instructionsForProg(p) {
			ins.validate(ctxt)
		}
	}
}

// assemble emits machine code.
// It is called at the very end of the assembly process.
func assemble(ctxt *obj.Link, cursym *obj.LSym, newprog obj.ProgAlloc) {
	if ctxt.Retpoline {
		ctxt.Diag("-spectre=ret not supported on %s", ctxt.Arch.Name)
		ctxt.Retpoline = false // don't keep printing
	}

	var symcode []uint32
	for p := cursym.Func().Text; p != nil; p = p.Link {
		for _, insn := range instructionsForProg(p) {
			if insn.length() == 0 {
				continue
			}

			ic, err := insn.encode()
			if err == nil {
				symcode = append(symcode, ic)
			}
		}
	}
	cursym.Size = int64(4 * len(symcode))

	cursym.Grow(cursym.Size)
	for p, i := cursym.P, 0; i < len(symcode); p, i = p[4:], i+1 {
		ctxt.Arch.ByteOrder.PutUint32(p, symcode[i])
	}

	obj.MarkUnsafePoints(ctxt, cursym.Func().Text, newprog, isUnsafePoint, nil)
}

func isUnsafePoint(p *obj.Prog) bool {
	if p.From.Reg == REG_TMP || p.To.Reg == REG_TMP {
		return true
	}
	for _, arg := range p.RestArgs {
		if arg.Reg == REG_TMP {
			return true
		}
	}
	return false
}
