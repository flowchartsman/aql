package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flowchartsman/aql/parser/ast"
)

type opValidator struct {
	err error
}

func newopValidator() *opValidator {
	return &opValidator{}
}

func (o *opValidator) Err() error {
	return o.err
}

func (o *opValidator) Visit(node ast.Node) ast.Visitor {
	if o.err != nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.ExprNode:
		checker := newopChecker(n)
		// global operation checks here
		checker.check(noDuplicates())
		// ensure types are correct
		switch n.Op {
		case ast.LT, ast.LTE, ast.GT, ast.GTE, ast.BET:
			checker.check(allNumericVals())
		case ast.SIM:
			checker.check(allStringVals())
		}
		// check between arity and ordering
		if n.Op == ast.BET {
			checker.check(
				validArity(2, 2),
				betweenOrdering(),
			)
		}
		if checker.err() != nil {
			o.err = checker.err()
		}
		return nil
	}
	return o
}

type opChecker struct {
	e    error
	node *ast.ExprNode
}

func newopChecker(node *ast.ExprNode) *opChecker {
	return &opChecker{
		node: node,
	}
}

func (o *opChecker) check(checks ...opCheck) {
	for _, c := range checks {
		if o.e == nil {
			if problem, argNum := c(o.node); problem != "" {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("expression [%s] operation error", o.node.FriendlyString()))
				if argNum >= 0 {
					sb.WriteString(fmt.Sprintf(" (value %d/%d)", argNum+1, len(o.node.RVals)))
				}
				sb.WriteString(fmt.Sprintf(": operation %s %s", o.node.Op, problem))
				o.e = errors.New(sb.String())
				// found an error, return
				return
			}
		}
	}
}

func (o *opChecker) err() error {
	return o.e
}

type opCheck func(node *ast.ExprNode) (problem string, argNum int)

func validArity(atleast, atmost int) opCheck {
	if atleast > atmost {
		panic("invalid arity check: atleast must be <= atmost")
	}
	return func(node *ast.ExprNode) (string, int) {
		if atleast == atmost && len(node.RVals) != atleast {
			return fmt.Sprintf("requires exactly %d arguments", atleast), -1
		}
		if len(node.RVals) < atleast {
			return fmt.Sprintf("requires at least %d arguments", atleast), -1
		}
		if len(node.RVals) > atleast {
			return fmt.Sprintf("exceeds the maximum of %d arguments", atmost), len(node.RVals) - 1
		}
		return "", 0
	}
}

func betweenOrdering() opCheck {
	return func(node *ast.ExprNode) (string, int) {
		if len(node.RVals) != 2 {
			panic("somehow between op with invalid arity slipped through")
		}
		var (
			l, r float64
		)
		switch lnv := node.RVals[0].(type) {
		case ast.IntVal:
			l = float64(lnv.Value())
		case ast.FloatVal:
			l = lnv.Value()
		}
		switch rnv := node.RVals[1].(type) {
		case ast.IntVal:
			r = float64(rnv.Value())
		case ast.FloatVal:
			r = rnv.Value()
		}
		if r <= l {
			return "requires the second value to be greater", 1
		}
		return "", 0
	}
}

func allNumericVals() opCheck {
	return func(node *ast.ExprNode) (string, int) {
		badArg := -1
		for i, rv := range node.RVals {
			if !valNumeric(rv) {
				if len(node.RVals) > 1 {
					badArg = i
				}
				return fmt.Sprintf("requires numeric arguments, found %s", rv.FriendlyType()), badArg
			}
		}
		return "", 0
	}
}

func allStringVals() opCheck {
	return func(node *ast.ExprNode) (string, int) {
		badArg := -1
		for i, rv := range node.RVals {
			if !valString(rv) {
				if len(node.RVals) > 1 {
					badArg = i
				}
				return fmt.Sprintf("requires string arguments, found %s", rv.FriendlyType()), badArg
			}
		}
		return "", 0
	}
}

func noDuplicates() opCheck {
	return func(node *ast.ExprNode) (string, int) {
		if len(node.RVals) <= 1 {
			return "", -1
		}
		found := map[string]struct{}{}
		for i, rv := range node.RVals {
			if _, ok := found[rv.ValStr()]; ok {
				return "duplicate argument found", i
			}
			found[rv.ValStr()] = struct{}{}
		}
		return "", 0
	}
}

func valNumeric(value ast.Val) bool {
	switch value.(type) {
	case ast.IntVal, ast.FloatVal, *ast.TimeVal:
		return true
	}
	return false
}

func valString(value ast.Val) bool {
	if _, ok := value.(ast.StringVal); ok {
		return true
	}
	return false
}
