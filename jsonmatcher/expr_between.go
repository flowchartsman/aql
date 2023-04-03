package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
)

func exprBetween(RVals []ast.Val) fieldExpr {
	if len(RVals) != 2 {
		// backstop
		panic(fmt.Sprintf("betweenMatcher expects two constant values - got %d", len(RVals)))
	}
	switch RVals[0].(type) {
	case *ast.FloatVal, *ast.IntVal:
		// numeric between
		var constantValues [2]float64

		switch v := RVals[0].(type) {
		case *ast.FloatVal:
			constantValues[0] = v.Value()
		case *ast.IntVal:
			constantValues[0] = float64(v.Value())
		}
		switch v := RVals[1].(type) {
		case *ast.FloatVal:
			constantValues[1] = v.Value()
		case *ast.IntVal:
			constantValues[1] = float64(v.Value())
		}

		return &exprFloat{
			values: constantValues,
			op:     ast.BET,
		}
	case *ast.TimeVal:
		// datetime between
		// 2nd argument guaranteed by validator
		return &exprDatetime{
			values: [2]int64{
				RVals[0].(*ast.TimeVal).Value().UnixNano(),
				RVals[1].(*ast.TimeVal).Value().UnixNano(),
			},
			op: ast.BET,
		}
	default:
		// backstop
		panic(fmt.Sprintf("bad value type for betweenMatcher: %T", RVals[0]))
	}
}
