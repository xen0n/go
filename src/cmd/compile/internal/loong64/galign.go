// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong64

import (
	"cmd/compile/internal/ssagen"
	"cmd/internal/obj/loong"
)

func Init(arch *ssagen.ArchInfo) {
	arch.LinkArch = &loong.Linkloong64

	arch.REGSP = loong.REG_SP
	arch.MAXWIDTH = 1 << 50

	arch.Ginsnop = ginsnop
	arch.Ginsnopdefer = ginsnop
	arch.ZeroRange = zeroRange

	arch.SSAMarkMoves = ssaMarkMoves
	arch.SSAGenValue = ssaGenValue
	arch.SSAGenBlock = ssaGenBlock
}
