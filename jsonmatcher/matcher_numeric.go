package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
)

func numericMatcher(op ast.Op, RVals []ast.Val) matcher {
	if len(RVals) != 1 {
		// backstop
		panic(fmt.Sprintf("numericMatcher expects only one constant value - got %d", len(RVals)))
	}
	switch v := RVals[0].(type) {
	case *ast.FloatVal:
		return &floatMatcher{
			values: [2]float64{v.Value()},
			op:     op,
		}
	case *ast.IntVal:
		return &floatMatcher{
			values: [2]float64{float64(v.Value())},
			op:     op,
		}
	case *ast.TimeVal:
		return &datetimeMatcher{
			values: [2]int64{v.Value().UnixNano()},
			op:     op,
		}
	default:
		// backstop
		panic(fmt.Sprintf("bad value type for numeric matcher: %T", RVals[0]))
	}
}
