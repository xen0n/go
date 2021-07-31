// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong

import "cmd/internal/obj"

// canonicalizeInsnArityForProg desugars one *obj.Prog representing an insn
// with one less operand than its expected arity, by duplicating the
// destination reg as source reg.
func canonicalizeInsnArityForProg(p *obj.Prog) {
	if isDirectiveInsn(p.As) || isPseudoInsn(p.As) {
		// Directives and pseudo-instructions are out-of-scope.
		return
	}

	enc, err := encodingForAs(p.As)
	if err != nil {
		// Error will be reported later, so just skip this step.
		return
	}

	insnArity := enc.fmt.arity()
	switch insnArity {
	case 0, 1:
		// Nothing to expand for these insn formats.

	case 2:
		if p.To.Type == obj.TYPE_NONE {
			p.To = p.From
		}

	default:
		expectedRestArgsLen := insnArity - 2
		if len(p.RestArgs) == expectedRestArgsLen-1 {
			// This is actually an append operation, contrary to
			// the name.
			p.SetRestArgs([]obj.Addr{p.To})
		}
	}
}
