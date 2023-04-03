package jsonmatcher

import (
	"github.com/flowchartsman/aql/parser"
	"github.com/valyala/fastjson"
)

// Matcher performs an AQL query against JSON to see if it matches
type Matcher struct {
	root     boolNode
	ppool    fastjson.ParserPool
	messages []*parser.ParserMessage
}

// NewMatcher creates a new matcher that returns whether a JSON document matches
// an AQL query
func NewMatcher(aqlQuery string /*options*/) (*Matcher, error) {
	visitor := parser.NewMessageVisitor(warningVisitor)
	root, err := parser.ParseQuery(aqlQuery, parser.Visitors(visitor))
	if err != nil {
		return nil, err
	}
	builder := newBuilder(true)

	return &Matcher{
		root:     builder.build(root),
		messages: visitor.Messages(),
	}, nil
}

// Match returns whether or not the query matches on a JSON document.
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
	return m.root.result(doc), nil
}

// MatchParsed allows matching on an already-parsed (and presumably validated)
// *fastjson.Value. It is the caller's responsibility to ensure that the parser
// is not re-used during this call.
func (m *Matcher) MatchParsed(doc *fastjson.Value) (bool, error) {
	return m.root.result(doc), nil
}

// Messages will return any hints or warning messages that the matcher may have
// generated during parsing
func (m *Matcher) Messages() []*parser.ParserMessage {
	return m.messages
}

func (m *Matcher) Stats() *MatchStats {
	return m.root.stats()
}

type MatcherOption func(*Matcher) error

func TrackQueryStats( /*options*/ ) MatcherOption {
	return func(m *Matcher) error {
		_ = m
		return nil
	}
}
