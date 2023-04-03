package parser

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
)

type exprCheck func(expr *ast.ExprNode) *ParseError

func opValidator(node ast.Node) error {
	switch n := node.(type) {
	case *ast.ExprNode:
		for _, check := range []exprCheck{
			checkValues,
			checkArity,
			checkRVals,
			checkBetween,
		} {
			if err := check(n); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkValues(e *ast.ExprNode) *ParseError {
	var failMsg string
	var badIdx int
	switch e.Op {
	case ast.LT, ast.LTE, ast.GT, ast.GTE, ast.BET:
		failMsg, badIdx = "needs numeric arguments", mustBeOneOf(e.RVals, ast.TypeInt, ast.TypeFloat, ast.TypeTime)
	case ast.SIM:
		// Temporarily accept regexp as well for legacy reasons. TODO: remove
		failMsg, badIdx = "needs string, or boolean arguments", mustBeOneOf(e.RVals, ast.TypeString, ast.TypeRegex, ast.TypeBool)
	default:
		return nil
	}
	if badIdx >= 0 {
		return ErrorWith(e.RVals[badIdx], fmt.Sprintf("[%s] operation %s", e.Op, failMsg))
	}
	return nil
}

func checkArity(e *ast.ExprNode) *ParseError {
	const (
		inf = -1
	)
	numArgs := len(e.RVals)
	var min, max int
	switch e.Op {
	// unary
	case ast.EXS, ast.NUL:
		min, max = 0, 0

	// binary
	case ast.LT, ast.LTE, ast.GT, ast.GTE:
		min, max = 1, 1

	// ternary
	case ast.BET:
		min, max = 2, 2

	// n-ary
	case ast.EQ, ast.SIM:
		min, max = 1, inf

	// backstop
	default:
		panic(fmt.Sprintf("undefined operation in arity check: [%s]", e.Op))
	}

	var msg string
	switch {
	case min == 0 && max == 0:
		if numArgs > 0 {
			msg = "does not accept arguments"
		}
	case min == max:
		if numArgs != min {
			msg = fmt.Sprintf("requires exactly %d arguments", min)
		}
	case min < max:
		if numArgs < min || numArgs > max {
			msg = fmt.Sprintf("requires between %d and %d arguments", min, max)
		}
	case max == inf:
		if min > 0 && numArgs < min {
			msg = fmt.Sprintf("requires at least %d arguments", min)
		}
	// backstop
	default:
		panic(fmt.Sprintf("invalid arity min/max: %d/%d", min, max))
	}

	if msg != "" {
		return ErrorWith(e, fmt.Sprintf("[%s] operation %s", e.Op, msg))
	}
	return nil
}

// check for duplicates and conflicting values
func checkRVals(e *ast.ExprNode) *ParseError {
	if len(e.RVals) <= 1 {
		return nil
	}
	seen := map[string]struct{}{}
	for i, rv := range e.RVals {
		sv := rv.String()
		if _, found := seen[sv]; found {
			return ErrorWith(e.RVals[i], fmt.Sprintf("duplicate argument [%s] (value %d/%d)", sv, i+1, len(e.RVals)))
		}
		if e.Op == ast.EQ && (sv == "true" || sv == "false") {
			conflictingBool := false
			if sv == "true" {
				_, conflictingBool = seen["false"]
			} else {
				_, conflictingBool = seen["true"]
			}
			if conflictingBool {
				return ErrorWith(e.RVals[i], fmt.Sprintf("conflicting boolean value [%s] (value %d/%d)", sv, i+1, len(e.RVals)))
			}
		}
		seen[rv.String()] = struct{}{}
	}
	return nil
}

func checkBetween(e *ast.ExprNode) *ParseError {
	if e.Op != ast.BET {
		return nil
	}
	if e.RVals[0].Type() == ast.TypeTime {
		if e.RVals[1].Type() != ast.TypeTime {
			return ErrorAt(e.RVals[1].Pos(), "second argument must also be a datetime value")
		}
		lt, rt := e.RVals[0].(*ast.TimeVal), e.RVals[1].(*ast.TimeVal)
		if lt.Value().After(rt.Value()) || lt.Value().Equal(rt.Value()) {
			return ErrorAt(e.RVals[1].Pos(), "[><] operation requires the second argument be greater")
		}
		return nil
	}

	var l, r float64
	switch lnv := e.RVals[0].(type) {
	case *ast.IntVal:
		l = float64(lnv.Value())
	case *ast.FloatVal:
		l = lnv.Value()
	}
	switch rnv := e.RVals[1].(type) {
	case *ast.IntVal:
		r = float64(rnv.Value())
	case *ast.FloatVal:
		r = rnv.Value()
	}
	if r <= l {
		return ErrorAt(e.RVals[1].Pos(), "[><] operation requires the second argument be greater")
	}
	return nil
}

func mustBeOneOf(values []ast.Val, types ...ast.ValType) (badIdx int) {
	if len(types) == 0 {
		panic("invalid type check")
	}
VLOOP:
	for v := range values {
		for t := range types {
			if values[v].Type() == types[t] {
				continue VLOOP
			}
		}
		return v
	}
	return -1
}
