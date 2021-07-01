package jsonquery

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/Jeffail/gabs/v2"
	"github.com/flowchartsman/aql/parser"
)

// Querier queries arbitrary json
type Querier struct {
	query      *parser.Node
	strictPath bool
}

// NewQuerier returns a new querier based on a parsed AQL query
func NewQuerier(query string, options ...QueryOption) (*Querier, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	rootnode, err := parser.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	q := &Querier{
		query: rootnode,
	}
	for _, o := range options {
		if err := o(q); err != nil {
			return nil, err
		}
	}
	return q, nil
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

func (q *Querier) MatchContainer(container *gabs.Container) (bool, error) {
	if container == nil {
		return false, fmt.Errorf("empty gabs container")
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
		rmatch, err := q.rdMatch(c, node.Right)
		if err != nil {
			return false, err
		}
		return lmatch && rmatch, nil
	case parser.NodeNot:
		// NOT nodes only have a LHS and return the negative of that match
		match, err := q.rdMatch(c, node.Left)
		if err != nil {
			return false, err
		}
		return !match, nil
	case parser.NodeTerminal:
		// hack for exists query for now.
		existsQuery := len(node.Comparison.Values) == 1 && node.Comparison.Values[0] == "exists"
		lvals, err := getLvals(node.Comparison.Field, c)
		if err != nil {
			if errors.Is(err, ErrPathNotFound) && (existsQuery || !q.strictPath) {
				return false, nil
			}
			return false, err
		}
		if existsQuery {
			return true, nil
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

var ErrPathNotFound = errors.New("path not found")

func getLvals(path []string, container *gabs.Container) ([]string, error) {
	target := container.S(path...)
	if target == nil {
		return nil, ErrPathNotFound
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
	// TODO: need to apply JSON types to lvals so that "true" != true and operators can pick their types
	for _, v := range flat {
		switch vv := v.(type) {
		case json.Number:
			lvals = append(lvals, vv.String())
		case bool:
			if vv {
				lvals = append(lvals, "true")
			} else {
				lvals = append(lvals, "false")
			}
		case string:
			lvals = append(lvals, vv)
		}
	}
	return lvals, nil
}

type QueryOption func(q *Querier) error

func StrictPath() QueryOption {
	return func(q *Querier) error {
		q.strictPath = true
		return nil
	}
}
