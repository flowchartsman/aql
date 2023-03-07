package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flowchartsman/aql/parser/ast"
)

type initializer struct {
	err error
}

func newInitializer() *initializer {
	return &initializer{}
}

func (in *initializer) Err() error {
	return in.err
}

func (in *initializer) Visit(node ast.Node) ast.Visitor {
	if in.err != nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.ExprNode:
		for i, v := range n.RVals {
			switch val := v.(type) {
			case *ast.RegexpVal:
				if _, err := val.Value(); err != nil {
					in.err = valueErr(n, i, "invalid regular expression", err)
					return nil
				}
			case *ast.NetVal:
				if _, err := val.Value(); err != nil {
					in.err = valueErr(n, i, "invalid net value", err)
					return nil
				}
			case *ast.TimeVal:
				if _, err := val.Value(); err != nil {
					in.err = valueErr(n, i, "invalid date value", err)
					return nil
				}
			}
		}
		return nil
	}
	return in
}

func valueErr(enode *ast.ExprNode, valNum int, desc string, err error) error {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("value error for expression [%s]", enode.FriendlyString()))
	if len(enode.RVals) > 1 {
		sb.WriteString(fmt.Sprintf(" value (%d/%d)", valNum+1, len(enode.RVals)))
	}
	sb.WriteString(fmt.Sprintf(": %s", desc))
	if err != nil {
		sb.WriteString(fmt.Sprintf(": %s", err))
	}
	return errors.New(sb.String())
}
