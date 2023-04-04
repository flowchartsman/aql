package jsonmatcher

type statsProvider interface {
	stats() *MatchStats
}

type boolNode interface {
	statsProvider
	// result() (bool, error)
	result([]byte) bool
}

type andNode struct {
	left, right boolNode
	nodeStats   *nodeStats
}

func (a *andNode) result(root []byte) bool {
	result := a.left.result(root) && a.right.result(root)
	if a.nodeStats != nil {
		a.nodeStats.mark(result)
	}
	return result
}

func (a *andNode) stats() *MatchStats {
	return a.nodeStats.toStatsNode()
}

type orNode struct {
	left, right boolNode
	nodeStats   *nodeStats
}

func (o *orNode) result(root []byte) bool {
	result := o.left.result(root) || o.right.result(root)
	if o.nodeStats != nil {
		o.nodeStats.mark(result)
	}
	return result
}

func (o *orNode) stats() *MatchStats {
	return o.nodeStats.toStatsNode()
}

type notNode struct {
	sub       boolNode
	nodeStats *nodeStats
}

func (n *notNode) result(root []byte) bool {
	result := !n.sub.result(root)
	if n.nodeStats != nil {
		n.nodeStats.mark(result)
	}
	return result
}

func (n *notNode) stats() *MatchStats {
	return n.nodeStats.toStatsNode()
}

// type subdocNode struct {
// 	prefix    []string
// 	sub       matcherNode
// 	nodeStats *nodeStats
// }

// func (s *subdocNode) matches(value *fastjson.Value) bool {
// 	rootDocs := getValues(s.prefix, value)
// 	matched := false
// 	for _, rootDoc := range rootDocs {
// 		if s.sub.matches(rootDoc) {
// 			matched = true
// 			break
// 		}
// 	}
// 	if s.nodeStats != nil {
// 		s.nodeStats.mark(matched)
// 	}
// 	return matched
// }

// func (s *subdocNode) stats() *StatsNode {
// 	if s.nodeStats == nil {
// 		return nil
// 	}
// 	return s.nodeStats.toStatsNode(s.sub)
// }

type exprNode struct {
	path      []string
	exprs     []fieldExpr
	nodeStats *nodeStats
}

func (e exprNode) result(root []byte) bool {
	matched := false
	field := getField(e.path, root)
	if len(field.values) > 0 {
		for _, m := range e.exprs {
			if m.matches(field) {
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

func (e exprNode) stats() *MatchStats {
	if e.nodeStats == nil {
		return nil
	}
	return e.nodeStats.toStatsNode()
}
