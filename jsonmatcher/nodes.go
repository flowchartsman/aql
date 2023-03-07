package jsonmatcher

import "github.com/valyala/fastjson"

type matcherNode interface {
	matches(value *fastjson.Value) bool
	stats() *StatsNode
}

type boolNode struct {
	and bool
	left,
	right matcherNode
	nodeStats *nodeStats
}

func (b *boolNode) matches(value *fastjson.Value) bool {
	matched := false
	if b.and {
		matched = b.left.matches(value) && b.right.matches(value)
	} else {
		matched = b.left.matches(value) || b.right.matches(value)
	}
	if b.nodeStats != nil {
		b.nodeStats.mark(matched)
	}
	return matched
}

func (b *boolNode) stats() *StatsNode {
	if b.nodeStats == nil {
		return nil
	}
	return b.nodeStats.toStatsNode(b.left, b.right)
}

type notNode struct {
	sub       matcherNode
	nodeStats *nodeStats
}

func (n *notNode) matches(value *fastjson.Value) bool {
	matched := !n.sub.matches(value)
	if n.nodeStats != nil {
		n.nodeStats.mark(matched)
	}
	return matched
}

func (n *notNode) stats() *StatsNode {
	if n.nodeStats == nil {
		return nil
	}
	return n.nodeStats.toStatsNode(n.sub)
}

type subdocNode struct {
	prefix    []string
	sub       matcherNode
	nodeStats *nodeStats
}

func (s *subdocNode) matches(value *fastjson.Value) bool {
	rootDocs := getValues(s.prefix, value)
	matched := false
	for _, rootDoc := range rootDocs {
		if s.sub.matches(rootDoc) {
			matched = true
			break
		}
	}
	if s.nodeStats != nil {
		s.nodeStats.mark(matched)
	}
	return matched
}

func (s *subdocNode) stats() *StatsNode {
	if s.nodeStats == nil {
		return nil
	}
	return s.nodeStats.toStatsNode(s.sub)
}

type exprNode struct {
	path      []string
	clauses   []clause
	nodeStats *nodeStats
}

func (e exprNode) matches(value *fastjson.Value) bool {
	matched := false
	testValues := getValues(e.path, value)
	if len(testValues) > 0 {
		for _, c := range e.clauses {
			if c.matches(testValues) {
				matched = true
				break
			}
		}
	}

	if e.nodeStats != nil {
		e.nodeStats.mark(matched)
	}
	return matched
}

func (e exprNode) stats() *StatsNode {
	if e.nodeStats == nil {
		return nil
	}
	return e.nodeStats.toStatsNode()
}
