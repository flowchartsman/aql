package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
)

// TODO: some optimizations possible here, including multi-term string matchers,
// which can be unified into a singular regular expression.
//
// i.e: combinedRVals := CoalesceRVals(n.RVals) for vtype, RVals := range
// combinedRVals
func exprEQ(RVals []ast.Val) []fieldExpr {
	if len(RVals) < 1 {
		// backstop
		panic("eqMatcher expects at least one constant value")
	}
	matchers := make([]fieldExpr, 0, len(RVals))
	for _, r := range RVals {
		switch rval := r.(type) {
		case *ast.StringVal:
			str := rval.Value()
			matchers = append(matchers, &exprRegexp{
				value: stringSearchRegexp(str),
			})
		case *ast.RegexpVal:
			matchers = append(matchers, &exprRegexp{
				value: rval.Value(),
			})
		case *ast.FloatVal:
			matchers = append(matchers, &exprFloat{
				values: [2]float64{rval.Value()},
				op:     ast.EQ,
			})
		case *ast.IntVal:
			matchers = append(matchers, &exprFloat{
				values: [2]float64{float64(rval.Value())},
				op:     ast.EQ,
			})
		case *ast.BoolVal:
			matchers = append(matchers, &exprEqBool{
				value: bool(rval.Value()),
				op:    ast.EQ,
			})
		case *ast.TimeVal:
			matchers = append(matchers, &exprDatetime{
				values: [2]int64{rval.Value().UnixNano()},
				op:     ast.EQ,
			})
		case *ast.NetVal:
			matchers = append(matchers, &exprNet{
				value: rval.Value(),
			})
		default:
			// backstop
			panic(fmt.Sprintf("bad value type for numeric matcher: %T", RVals[0]))
		}
	}
	return matchers
}
