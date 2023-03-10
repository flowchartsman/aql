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

func (n *numericClause) matches(values []*fastjson.Value) bool {
	for _, v := range values {
		fv, ok := getNumberVal(v)
		if !ok {
			continue
		}
		switch n.op {
		case ast.EQ:
			if fv == n.value {
				return true
			}
		case ast.LT:
			if fv < n.value {
				return true
			}
		case ast.LTE:
			if fv <= n.value {
				return true
			}
		case ast.GT:
			if fv > n.value {
				return true
			}
		case ast.GTE:
			if fv >= n.value {
				return true
			}
		default:
			panic(fmt.Sprintf("invalid op for numeric comparison: %s", n.op))
		}
	}
	return false
}
