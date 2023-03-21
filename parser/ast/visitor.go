package ast

// Visitor's Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(node Node) (w Visitor)
}

// Walk traverses an AST in depth-first order: It starts by calling
// v.Visit(node); node must not be nil. If the visitor w returned by
// v.Visit(node) is not nil, Walk is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).
//
// TODO, allow visitor injection to avoid walking the tree multiple times.
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}
	switch n := node.(type) {
	case *AndNode:
		Walk(v, n.Left)
		Walk(v, n.Right)
	case *OrNode:
		Walk(v, n.Left)
		Walk(v, n.Right)
	case *NotNode:
		Walk(v, n.Expr)
	case *SubdocNode:
		Walk(v, n.Expr)
	case *ExprNode:
		Walk(v, n)
	}
	v.Visit(nil)
}
