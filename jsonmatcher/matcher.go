package jsonmatcher

import (
	"github.com/flowchartsman/aql/parser"
	"github.com/valyala/fastjson"
)

// Matcher performs an AQL query against JSON to see if it matches
type Matcher struct {
	root  matcherNode
	ppool fastjson.ParserPool
	// fieldstats map[string]FieldStats
}

func NewMatcher(aqlQuery string /*options*/) (*Matcher, error) {
	p := parser.NewParser(parser.InitGoTypes())
	root, err := p.ParseQuery(aqlQuery)
	if err != nil {
		return nil, err
	}
	builder := newBuilder(true)
	return &Matcher{
		root: builder.build(root),
	}, nil
}

func (m *Matcher) Match(json []byte) (bool, error) {
	// An unfortunate necessity for now, until something like
	// https://github.com/valyala/fastjson/pull/68 can land.
	if err := fastjson.ValidateBytes(json); err != nil {
		return false, err
	}
	parser := m.ppool.Get()
	defer m.ppool.Put(parser)
	doc, err := parser.ParseBytes(json)
	if err != nil {
		return false, err
	}
	return m.root.matches(doc), nil
}

func (m *Matcher) Stats() *StatsNode {
	return m.root.stats()
}

type MatcherOption func(*Matcher) error

func TrackQueryStats( /*options*/ ) MatcherOption {
	return func(m *Matcher) error {
		_ = m
		return nil
	}
}
