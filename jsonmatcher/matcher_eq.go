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
func eqMatcher(RVals []ast.Val) []matcher {
	if len(RVals) < 1 {
		// backstop
		panic("eqMatcher expects at least one constant value")
	}
	matchers := make([]matcher, 0, len(RVals))
	for _, r := range RVals {
		switch rval := r.(type) {
		case *ast.StringVal:
			str := rval.Value()
			matchers = append(matchers, &regexpMatcher{
				value: stringSearchRegexp(str),
			})
		case *ast.RegexpVal:
			matchers = append(matchers, &regexpMatcher{
				value: rval.Value(),
			})
		case *ast.FloatVal:
			matchers = append(matchers, &floatMatcher{
				values: [2]float64{rval.Value()},
				op:     ast.EQ,
			})
		case *ast.IntVal:
			matchers = append(matchers, &floatMatcher{
				values: [2]float64{float64(rval.Value())},
				op:     ast.EQ,
			})
		case *ast.BoolVal:
			matchers = append(matchers, &boolMatcher{
				value: bool(rval.Value()),
				op:    ast.EQ,
			})
		case *ast.TimeVal:
			matchers = append(matchers, &datetimeMatcher{
				values: [2]int64{rval.Value().UnixNano()},
				op:     ast.EQ,
			})
		case *ast.NetVal:
			matchers = append(matchers, &netMatcher{
				value: rval.Value(),
			})
		default:
			// backstop
			panic(fmt.Sprintf("bad value type for numeric matcher: %T", RVals[0]))
		}
	}
	return matchers
}
