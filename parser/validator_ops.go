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
			checkNoDuplicates,
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
		failMsg, badIdx = "needs string arguments", mustBeOneOf(e.RVals, ast.TypeString, ast.TypeRegex)
	default:
		return nil
	}
	if badIdx >= 0 {
		return ErrorAt(e.RVals[badIdx].Pos(), fmt.Sprintf("[%s] operation %s", e.Op, failMsg))
	}
	return nil
}

func checkArity(e *ast.ExprNode) *ParseError {
	numArgs := len(e.RVals)
	var min, max int
	switch e.Op {
	case ast.LT, ast.LTE, ast.GT, ast.GTE:
		min, max = 1, 1
	case ast.BET:
		min, max = 2, 2
	case ast.EXS, ast.NUL:
		min, max = 0, 0
	default:
		min, max = 1, 0
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
	case min > max:
		if numArgs < min {
			msg = fmt.Sprintf("requires at least %d arguments", min)
		}
	}
	if msg != "" {
		return ErrorAt(e.Pos(), fmt.Sprintf("[%s] operation %s", e.Op, msg))
	}
	return nil
}

func checkNoDuplicates(e *ast.ExprNode) *ParseError {
	if len(e.RVals) < 2 {
		return nil
	}
	found := map[string]struct{}{}
	for i, rv := range e.RVals {
		if _, ok := found[rv.String()]; ok {
			return ErrorAt(e.RVals[i].Pos(), fmt.Sprintf("duplicate argument [%s] (value %d/%d)", rv.String(), i+1, len(e.RVals)))
		}
		found[rv.String()] = struct{}{}
	}
	return nil
}

func checkBetween(e *ast.ExprNode) *ParseError {
	if e.Op != ast.BET {
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
