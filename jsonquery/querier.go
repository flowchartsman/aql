package jsonquery

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Jeffail/gabs/v2"
	"github.com/flowchartsman/aql/parser"
)

// Querier queries arbitrary json
type Querier struct {
	query *parser.Node
}

// NewQuerier returns a new querier based on a parsed AQL query
func NewQuerier(query string) (*Querier, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	rootnode, err := parser.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	return &Querier{
		query: rootnode,
	}, nil
}

func (q *Querier) Match(jsonData io.Reader) (bool, error) {
	dec := json.NewDecoder(jsonData)
	dec.UseNumber()
	container, err := gabs.ParseJSONDecoder(dec)
	if err != nil {
		return false, fmt.Errorf("error parsing json: %w", err)
	}
	return q.rdMatch(container, q.query)
}

func (q *Querier) rdMatch(c *gabs.Container, node *parser.Node) (bool, error) {
	switch node.NodeType {
	case parser.NodeOr:
		match, err := q.rdMatch(c, node.Left)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
		match, err = q.rdMatch(c, node.Right)
		if err != nil {
			return false, err
		}
		return match, nil
	case parser.NodeAnd:
		lmatch, err := q.rdMatch(c, node.Left)
		if err != nil {
			return false, err
		}
		rmatch, err := q.rdMatch(c, node.Left)
		if err != nil {
			return false, err
		}
		return lmatch && rmatch, nil
	case parser.NodeTerminal:
		lvals, err := getLvals(node.Comparison.Field, c)
		if err != nil {
			return false, err
		}
		// TODO premake!
		operation, ok := operations[node.Comparison.Op]
		if !ok {
			return false, fmt.Errorf("unknown operation %q", node.Comparison.Op)
		}
		comparator, err := operation.GetComparator(node.Comparison.Values)
		if err != nil {
			return false, err
		}
		return comparator.Evaluate(lvals)
	default:
		return false, fmt.Errorf("bad node type %d", node.NodeType)
	}
}

func getLvals(path []string, container *gabs.Container) ([]string, error) {
	target := container.S(path...)
	if target == nil {
		return nil, fmt.Errorf("path not found")
	}

	flat, err := target.Flatten()
	switch err {
	case gabs.ErrNotObjOrArray:
		flat = map[string]interface{}{"": target.Data()}
	case nil:
		// proceed
	default:
		return nil, fmt.Errorf("error getting query values: %w", err)
	}

	lvals := make([]string, 0, len(flat))
	for _, v := range flat {
		switch vv := v.(type) {
		case json.Number:
			lvals = append(lvals, vv.String())
		case string:
			lvals = append(lvals, vv)
		}
	}
	return lvals, nil
}
