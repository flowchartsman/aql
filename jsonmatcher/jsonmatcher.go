package jsonmatcher

import (
	"encoding/json"
	"errors"

	"github.com/flowchartsman/aql/parser"
)

// Matcher performs an AQL query against JSON to see if it matches
type Matcher struct {
	root     boolNode
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
func (m *Matcher) Match(data []byte) (bool, error) {
	if !json.Valid(data) {
		return false, errors.New("invalid JSON")
	}
	return m.root.result(data), nil
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
