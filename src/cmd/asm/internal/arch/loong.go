// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file encapsulates some of the odd characteristics of the LoongArch
// instruction set, to minimize its interaction with the core of the
// assembler.

package arch

import (
	"cmd/internal/obj"
	"cmd/internal/obj/loong"
)

// IsLoongBinaryDualInputs reports whether the op (as defined by a loong.A*
// constant) is one of the two-operand instructions whose both operands are
// inputs.
func IsLoongBinaryDualInputs(op obj.As) bool {
	switch op {
	case loong.AASRTLE, loong.AASRTGT, loong.ALDPTE:
		return true
	}
	return false
}

// IsLoongBinaryDualOutputs reports whether the op (as defined by a loong.A*
// constant) is one of the two-operand instructions whose both operands are
// outputs.
func IsLoongBinaryDualOutputs(op obj.As) bool {
	switch op {
	case loong.ARDTICKLW, loong.ARDTICKHW, loong.ARDTICKD:
		return true
	}
	return false
}

// IsLoongTernaryAllInputs reports whether the op (as defined by a loong.A*
// constant) is one on the three-operand instructions whose all operands are
// inputs.
func IsLoongTernaryAllInputs(op obj.As) bool {
	switch op {
	case loong.ACACOP, loong.ATLBINV, loong.APRELD, loong.APRELDX:
		return true
	}
	return false
}
