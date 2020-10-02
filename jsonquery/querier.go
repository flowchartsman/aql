package jsonquery

import (
	"fmt"

	"github.com/flowchartsman/aql/parser"
	"github.com/valyala/fastjson"
)

// Querier queries arbitrary json
type Querier struct {
	pp fastjson.ParserPool
}

func (q *Querier) Match(jsonData []byte, query *parser.Node) (bool, error) {
	p := q.pp.Get()
	defer q.pp.Put(p)
	return q.rdMatch(p, query)
}

func (q *Querier) rdMatch(p *fastjson.Parser, node *parser.Node) (bool, error) {
	switch node.NodeType {
	case parser.NodeOr:
		match, err := q.rdMatch(p, node.Left)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
		match, err = q.rdMatch(p, node.Right)
		if err != nil {
			return false, err
		}
		return match, nil
	case parser.NodeAnd:
		lmatch, err := q.rdMatch(p, node.Left)
		if err != nil {
			return false, err
		}
		rmatch, err := q.rdMatch(p, node.Left)
		if err != nil {
			return false, err
		}
		return lmatch && rmatch, nil
	case parser.NodeTerminal:
		//terminal matching below
	default:
		return false, fmt.Errorf("bad node type %d", node.NodeType)
	}
	return false, nil
}
