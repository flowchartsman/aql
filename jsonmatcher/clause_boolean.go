package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
	"github.com/valyala/fastjson"
)

type boolClause struct {
	value bool
	op    ast.Op
}

func (s *boolClause) matches(values []*fastjson.Value) bool {
	for _, v := range values {
		var (
			bv bool
			ok bool
		)
		switch s.op {
		case ast.EQ:
			bv, ok = getBoolVal(v)
		case ast.SIM:
			bv, ok = getTruthyVal(v)
		default:
			panic(fmt.Sprintf("invalid op for boolean comparison: %s", s.op))
		}
		if !ok {
			continue
		}
		if bv == s.value {
			return true
		}
	}
	return false
}
