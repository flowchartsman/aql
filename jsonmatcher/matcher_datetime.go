package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
	"github.com/valyala/fastjson"
)

type datetimeMatcher struct {
	values [2]int64
	op     ast.Op
}

func (m *datetimeMatcher) matches(values []*fastjson.Value) bool {
	for _, v := range values {
		dv, ok := getDatetimeVal(v)
		if !ok {
			continue
		}
		switch m.op {
		case ast.EQ:
			if dv == m.values[0] {
				return true
			}
		case ast.LT:
			if dv < m.values[0] {
				return true
			}
		case ast.LTE:
			if dv <= m.values[0] {
				return true
			}
		case ast.GT:
			if dv > m.values[0] {
				return true
			}
		case ast.GTE:
			if dv >= m.values[0] {
				return true
			}
		case ast.BET:
			if dv >= m.values[0] && dv <= m.values[1] {
				return true
			}
		// backstop
		default:
			panic(fmt.Sprintf("invalid op for datetime comparison: %s", m.op))
		}
	}
	return false
}
