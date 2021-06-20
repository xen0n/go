// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong

import (
	"cmd/internal/obj"
	"fmt"
)

func init() {
	obj.RegisterRegister(obj.RBaseLoong, REG_END, RegName)
	obj.RegisterOpcode(obj.ABaseLoong, Anames)
}

func RegName(r int) string {
	switch {
	case r == 0:
		return "NONE"
	case r == REG_G:
		return "g"
	case r == REG_SP:
		return "SP"
	case REG_R0 <= r && r <= REG_R31:
		return fmt.Sprintf("R%d", r-REG_R0)
	case REG_F0 <= r && r <= REG_F31:
		return fmt.Sprintf("F%d", r-REG_F0)
	case REG_FCC0 <= r && r <= REG_FCC7:
		return fmt.Sprintf("FCC%d", r-REG_FCC0)
	default:
		return fmt.Sprintf("Rgok(%d)", r-obj.RBaseLoong)
	}
}
