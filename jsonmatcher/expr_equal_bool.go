package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
)

type exprEqBool struct {
	value bool
	op    ast.Op
}

func (e *exprEqBool) matches(field *field) bool {
	for _, v := range field.scalarValues() {
		var (
			bv bool
			ok bool
		)
		switch e.op {
		case ast.EQ:
			bv, ok = getBoolVal(v)
		case ast.SIM:
			bv, ok = getTruthyVal(v)
		default:
			// backstop
			panic(fmt.Sprintf("invalid op for boolean comparison: %s", e.op))
		}
		if !ok {
			continue
		}
		if bv == e.value {
			return true
		}
	}
	return false
}
