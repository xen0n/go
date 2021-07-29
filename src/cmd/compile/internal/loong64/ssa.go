// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong64

import (
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/ssagen"
	"cmd/internal/obj"
)

// markMoves marks any MOVXconst ops that need to avoid clobbering flags.
// LoongArch has no flags, so this is a no-op.
func ssaMarkMoves(s *ssagen.State, b *ssa.Block) {}

func ssaGenValue(s *ssagen.State, v *ssa.Value) {
	s.SetPos(v.Pos)

	switch v.Op {
	case ssa.OpLoong64ADDD:
		out1 := v.Reg()
		in1 := v.Args[0].Reg()
		in2 := v.Args[1].Reg()
		p := s.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_REG
		p.From.Reg = in1
		p.SetRestArgs([]obj.Addr{
			{Type: obj.TYPE_REG, Reg: in2},
		})
		p.To.Type = obj.TYPE_REG
		p.To.Reg = out1

	default:
		v.Fatalf("Unhandled op %v", v.Op)
	}
}

func ssaGenBlock(s *ssagen.State, b, next *ssa.Block) {
	s.SetPos(b.Pos)

	switch b.Kind {
	default:
		b.Fatalf("Unhandled block: %s", b.LongString())
	}
}
