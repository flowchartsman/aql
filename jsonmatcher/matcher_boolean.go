package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
	"github.com/valyala/fastjson"
)

type boolMatcher struct {
	value bool
	op    ast.Op
}

func (s *boolMatcher) matches(values []*fastjson.Value) bool {
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
			// backstop
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
