package jsonmatcher

import (
	"github.com/flowchartsman/aql/parser/ast"
)

type builder struct {
	withStats bool
	// track all expression node fields for returning detailed stats on what the
	// query is encountering in the field
	// integrate node_stats
	// tracks current subdoc prefix so they can be nested
	// subdocPrefix []string
}

func newBuilder(withStats bool) *builder {
	b := &builder{
		withStats: withStats,
	}
	if withStats {
		// integrate field_stats
	}
	return b
}

func (b *builder) build(node ast.Node) matcherNode {
	// make a set of expectedtypes during build to build fieldstats
	var matcher matcherNode
	switch n := node.(type) {
	case *ast.AndNode:
		node := &boolNode{
			and:   true,
			left:  b.build(n.Left),
			right: b.build(n.Right),
		}
		if b.withStats {
			node.nodeStats = &nodeStats{
				nodeName: "AND",
			}
		}
		matcher = node
	case *ast.OrNode:
		node := &boolNode{
			and:   false,
			left:  b.build(n.Left),
			right: b.build(n.Right),
		}
		if b.withStats {
			node.nodeStats = &nodeStats{
				nodeName: "OR",
			}
		}
		matcher = node
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
		// TODO: several optimizations possible here, including multi-term
		// string matchers, which can be unified into a singular regular
		// expression.
		// combinedRvals := CoalesceRvals(n.RVals)
		// for vtype, RVals := range combinedRvals
		clauses := make([]clause, 0, len(n.RVals))
		for _, r := range n.RVals {
			switch rval := r.(type) {
			case ast.StringVal:
				str := rval.Value()
				// switch n.Op {
				// case ast.EQ:
				// 	clauses = append(clauses, &regexpClause{
				// 		value: stringSearchRegexp(str),
				// 	})
				// case ast.SIM:
				// 	clauses = append(clauses, &fuzzyClause{

				// 	})
				// }
				clauses = append(clauses, &regexpClause{
					value: stringSearchRegexp(str),
				})
			case *ast.RegexpVal:
				switch n.Op {
				// TODO: remove SIM, eventually
				case ast.EQ, ast.SIM:
					// don't need to check for error here, since regular
					// expression will be initialized
					// TODO: remove error check entirely in favor of error
					// thrown from InitGoTypes
					rex, _ := rval.Value()
					clauses = append(clauses, &regexpClause{
						value: rex,
					})
				}
			case ast.FloatVal:
				clauses = append(clauses, &numericClause{
					value: rval.Value(),
					op:    n.Op,
				})
			case ast.IntVal:
				clauses = append(clauses, &numericClause{
					value: float64(rval.Value()),
					op:    n.Op,
				})
			case ast.BoolVal:
				clauses = append(clauses, &boolClause{
					value: bool(rval.Value()),
					op:    n.Op,
				})
			case *ast.TimeVal:
				tval, _ := rval.Value()
				clauses = append(clauses, &datetimeClause{
					value: tval,
					op:    n.Op,
				})
			case *ast.NetVal:
				nval, _ := rval.Value()
				clauses = append(clauses, &netClause{
					value: nval,
					op:    n.Op,
				})
			default:
				// invalid
			}
		}
		node.clauses = clauses
		matcher = node
	}
	return matcher
}
