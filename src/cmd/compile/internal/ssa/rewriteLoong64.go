// Code generated from gen/Loong64.rules; DO NOT EDIT.
// generated with: cd gen; go run *.go

package ssa

func rewriteValueLoong64(v *Value) bool {
	switch v.Op {
	case OpAdd64:
		v.Op = OpLoong64ADDD
		return true
	case OpNilCheck:
		v.Op = OpLoong64ADDD
		return true
	}
	return false
}
func rewriteBlockLoong64(b *Block) bool {
	return false
}
