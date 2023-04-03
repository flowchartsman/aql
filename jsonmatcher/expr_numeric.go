package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
)

func exprNumeric(op ast.Op, RVals []ast.Val) fieldExpr {
	if len(RVals) != 1 {
		// backstop
		panic(fmt.Sprintf("numericMatcher expects only one constant value - got %d", len(RVals)))
	}
	switch v := RVals[0].(type) {
	case *ast.FloatVal:
		return &exprFloat{
			values: [2]float64{v.Value()},
			op:     op,
		}
	case *ast.IntVal:
		return &exprFloat{
			values: [2]float64{float64(v.Value())},
			op:     op,
		}
	case *ast.TimeVal:
		return &exprDatetime{
			values: [2]int64{v.Value().UnixNano()},
			op:     op,
		}
	default:
		// backstop
		panic(fmt.Sprintf("bad value type for numeric matcher: %T", RVals[0]))
	}
}
