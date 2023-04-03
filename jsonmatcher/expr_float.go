package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
)

type exprFloat struct {
	values [2]float64
	op     ast.Op
}

func (e *exprFloat) matches(field *field) bool {
	for _, v := range field.scalarValues() {
		fv, ok := getNumberVal(v)
		if !ok {
			continue
		}
		switch e.op {
		case ast.EQ:
			if fv == e.values[0] {
				return true
			}
		case ast.LT:
			if fv < e.values[0] {
				return true
			}
		case ast.LTE:
			if fv <= e.values[0] {
				return true
			}
		case ast.GT:
			if fv > e.values[0] {
				return true
			}
		case ast.GTE:
			if fv >= e.values[0] {
				return true
			}
		case ast.BET:
			if fv >= e.values[0] && fv <= e.values[1] {
				return true
			}
		// backstop
		default:
			panic(fmt.Sprintf("invalid op for numeric comparison: %s", e.op))
		}
	}
	return false
}
