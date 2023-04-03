package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
)

type exprDatetime struct {
	values [2]int64
	op     ast.Op
}

func (e *exprDatetime) matches(field *field) bool {
	for _, v := range field.scalarValues() {
		dv, ok := getDatetimeVal(v)
		if !ok {
			continue
		}
		switch e.op {
		case ast.EQ:
			if dv == e.values[0] {
				return true
			}
		case ast.LT:
			if dv < e.values[0] {
				return true
			}
		case ast.LTE:
			if dv <= e.values[0] {
				return true
			}
		case ast.GT:
			if dv > e.values[0] {
				return true
			}
		case ast.GTE:
			if dv >= e.values[0] {
				return true
			}
		case ast.BET:
			if dv >= e.values[0] && dv <= e.values[1] {
				return true
			}
		// backstop
		default:
			panic(fmt.Sprintf("invalid op for datetime comparison: %s", e.op))
		}
	}
	return false
}
