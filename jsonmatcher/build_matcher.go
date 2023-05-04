package jsonmatcher

import (
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
)

type builder struct {
	withStats bool
	// track all expression node fields for returning detailed stats on what the
	// query is encountering in the field
	// integrate node_stats
}

func newBuilder(withStats bool) *builder {
	b := &builder{
		withStats: withStats,
	}
	// TODO: Fieldstats (optional)
	// if withStats {
	// 	integrate field_stats
	// }
	return b
}

func (b *builder) build(node ast.Node) boolNode {
	// make a set of expectedtypes during build to build fieldstats
	// var matcher matcherNode
	switch n := node.(type) {
	case *ast.AndNode:
		node := &andNode{
			left:  b.build(n.Left),
			right: b.build(n.Right),
		}
		if b.withStats {
			node.nodeStats = &nodeStats{
				nodeName: "AND",
			}
		}
		return node
	case *ast.OrNode:
		node := &orNode{
			left:  b.build(n.Left),
			right: b.build(n.Right),
		}
		if b.withStats {
			node.nodeStats = &nodeStats{
				nodeName: "OR",
			}
		}
		return node
	case *ast.NotNode:
		node := &notNode{
			sub: b.build(n.Expr),
		}
		if b.withStats {
			node.nodeStats = &nodeStats{
				nodeName: "NOT",
			}
		}
		return node
	// case *ast.SubdocNode:
	// tracks current subdoc prefix so they can be nested
	// subdocPrefix []string
	// 	b.subdocPrefix = append(b.subdocPrefix, n.Field...)
	// 	node := &subdocNode{
	// 		prefix: b.subdocPrefix,
	// 		sub:    b.build(n.Expr),
	// 	}
	// 	if b.withStats {
	// 		node.nodeStats = &nodeStats{
	// 			nodeName: ast.FieldString(n.Field) + "{}",
	// 		}
	// 	}
	// 	b.subdocPrefix = b.subdocPrefix[:len(b.subdocPrefix)-len(n.Field)]
	// 	return node
	case *ast.ExprNode:
		node := &exprNode{
			path: n.Field,
		}
		if b.withStats {
			node.nodeStats = &nodeStats{
				nodeName: n.FriendlyString(),
			}
		}
		// would assign expected types here
		// TODO: single matcher and unify
		// closures?
		switch n.Op {
		// unary
		case ast.EXS:
			node.exprs = []fieldExpr{
				&exprExists{},
			}
		case ast.NUL:
			node.exprs = []fieldExpr{
				&exprNull{},
			}
		// binary:
		case ast.LT, ast.LTE, ast.GT, ast.GTE:
			node.exprs = []fieldExpr{
				exprNumeric(n.Op, n.RVals),
			}
		// ternary
		case ast.BET:
			node.exprs = []fieldExpr{
				exprBetween(n.RVals),
			}
		// n-ary
		case ast.EQ:
			node.exprs = exprEQ(n.RVals)
		case ast.SIM:
			node.exprs = exprEQ(n.RVals)
			// exprSim Deprecated
			// node.exprs = exprSim(n.RVals)
		default:
			// backstop
			panic(fmt.Sprintf("undefined operation: [%s]", n.Op))
		}
		return node
	}
	// return matcher
	return nil
}
