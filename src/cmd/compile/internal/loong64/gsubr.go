// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong64

import (
	"cmd/compile/internal/objw"
	"cmd/internal/obj"
	"cmd/internal/obj/loong"
)

func ginsnop(pp *objw.Progs) *obj.Prog {
	// Hardware nop is ANDI $0, ZERO, ZERO
	p := pp.Prog(loong.AANDI)
	p.From.Type = obj.TYPE_CONST
	p.SetRestArgs([]obj.Addr{{Type: obj.TYPE_REG, Reg: loong.REG_ZERO}})
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: loong.REG_ZERO}
	return p
}
