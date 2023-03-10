package jsonmatcher

import (
	"github.com/flowchartsman/aql/parser/ast"
)

type builder struct {
	withStats bool
	// track all expression node fields for returning detailed stats on what the
	// query is encountering in the field
	fieldsExpect *expectedTypeSet
	// tracks current subdoc prefix so they can be nested
	subdocPrefix []string
}

func newBuilder(withStats bool) *builder {
	b := &builder{
		withStats: withStats,
	}
	if withStats {
		b.fieldsExpect = newExpectedTypeSet()
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
	case *ast.SubdocNode:
		b.subdocPrefix = append(b.subdocPrefix, n.Field...)
		node := &subdocNode{
			prefix: b.subdocPrefix,
			sub:    b.build(n.Expr),
		}
		if b.withStats {
			node.nodeStats = &nodeStats{
				nodeName: ast.FieldString(n.Field) + "{}",
			}
		}
		b.subdocPrefix = b.subdocPrefix[:len(b.subdocPrefix)-len(n.Field)]
		return node
	case *ast.ExprNode:
		node := &exprNode{
			path: n.Field,
		}
		if b.withStats {
			node.nodeStats = &nodeStats{
				nodeName: n.FriendlyString(),
			}
		}
		clauses := make([]clause, 0, len(n.RVals))
		for _, r := range n.RVals {
			switch rval := r.(type) {
			// case ast.ExactStringVal:
			// EQ is exact match only, so stringclause
			// SIM is case-sensitive wordregexp
			case ast.StringVal:
				switch n.Op {
				case ast.EQ:
					clauses = append(clauses, &regexpClause{
						value: wordRegexp(rval.Value()),
					})
				case ast.SIM:
					clauses = append(clauses, &regexpClause{
						value: wildcardRegexp(rval.Value()),
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
			}
		}
		node.clauses = clauses
		matcher = node
	}
	return matcher
}

// for gathering what types different fields are expecting (for field-based statistics)
type expectedTypeSet struct {
	fieldExpects map[string]map[expectedType]struct{}
}

func newExpectedTypeSet() *expectedTypeSet {
	return &expectedTypeSet{
		fieldExpects: map[string]map[expectedType]struct{}{},
	}
}

func (es *expectedTypeSet) addExpected(field []string, expected expectedType) {
	fieldStr := ast.FieldString(field)
	if es.fieldExpects[fieldStr] == nil {
		es.fieldExpects[fieldStr] = map[expectedType]struct{}{}
	}
	es.fieldExpects[fieldStr][expected] = struct{}{}
}

func (ex *expectedTypeSet) getFields() map[string][]expectedType {
	out := map[string][]expectedType{}
	for field, fieldExpects := range ex.fieldExpects {
		outfield := []expectedType{}
		for e := range fieldExpects {
			outfield = append(outfield, e)
		}
		out[field] = outfield
	}
	return out
}
