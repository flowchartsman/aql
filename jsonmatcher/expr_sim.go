package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
)

func exprSim(RVals []ast.Val) []fieldExpr {
	if len(RVals) < 1 {
		// backstop
		panic("simMatcher expects at least one constant value")
	}
	matchers := make([]fieldExpr, 0, len(RVals))
	for _, r := range RVals {
		switch rval := r.(type) {
		case *ast.StringVal:
			matchers = append(matchers, newExprFuzzy(rval.Value()))
		case *ast.BoolVal:
			matchers = append(matchers, &exprEqBool{
				value: bool(rval.Value()),
				op:    ast.SIM,
			})
		// same as EQ for legacy reasons. TODO: remove
		case *ast.RegexpVal:
			matchers = append(matchers, &exprRegexp{
				value: rval.Value(),
			})
		default:
			// backstop
			panic(fmt.Sprintf("bad value type for similarity matcher: %T", RVals[0]))
		}
	}
	return matchers
}
