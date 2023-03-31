package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
	"github.com/valyala/fastjson"
)

type floatMatcher struct {
	values [2]float64
	op     ast.Op
}

func (m *floatMatcher) matches(values []*fastjson.Value) bool {
	for _, v := range values {
		fv, ok := getNumberVal(v)
		if !ok {
			continue
		}
		switch m.op {
		case ast.EQ:
			if fv == m.values[0] {
				return true
			}
		case ast.LT:
			if fv < m.values[0] {
				return true
			}
		case ast.LTE:
			if fv <= m.values[0] {
				return true
			}
		case ast.GT:
			if fv > m.values[0] {
				return true
			}
		case ast.GTE:
			if fv >= m.values[0] {
				return true
			}
		case ast.BET:
			if fv >= m.values[0] && fv <= m.values[1] {
				return true
			}
		// backstop
		default:
			panic(fmt.Sprintf("invalid op for numeric comparison: %s", m.op))
		}
	}
	return false
}
