package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
	"github.com/valyala/fastjson"
)

type numericClause struct {
	value float64
	op    ast.Op
}

func (s *numericClause) matches(values []*fastjson.Value) bool {
	for _, v := range values {
		fv, ok := getNumberVal(v)
		if !ok {
			continue
		}
		switch s.op {
		case ast.EQ:
			if fv == s.value {
				return true
			}
		case ast.LT:
			if fv < s.value {
				return true
			}
		case ast.LTE:
			if fv <= s.value {
				return true
			}
		case ast.GT:
			if fv > s.value {
				return true
			}
		case ast.GTE:
			if fv >= s.value {
				return true
			}
		default:
			panic(fmt.Sprintf("invalid op for numeric comparison: %s", s.op))
		}
	}
	return false
}
