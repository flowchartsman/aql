package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/flowchartsman/aql/internal/grammar"
	"github.com/flowchartsman/aql/parser/ast"
)

// ParseError is a detailed parser error.
// It's a type alias to allow exporting the internal type from the generated code.
type ParseError = grammar.ParseError

func ParseQuery(query string, options ...Option) (ast.Node, error) {
	return ParseQueryReader(strings.NewReader(query), options...)
}

func ParseQueryReader(r io.Reader, options ...Option) (ast.Node, error) {
	opts := &ParserOpts{
		visitors: []Visitor{
			VisitorFunc(opValidator),
		},
	}
	for _, o := range options {
		o(opts)
	}
	v, err := grammar.ParseReader("", r, grammar.Debug(opts.debug))
	if err != nil {
		return nil, grammar.GetParseError(err)
	}

	var root ast.Node

	switch t := v.(type) {
	case nil:
		return nil, genericParseError("parser returned nil output")
	case ast.Node:
		root = t
	default:
		return nil, genericParseError(fmt.Sprintf("parser returned unknown type: %T", t))
	}

	// TODO Accumulate errors/messages from here on out and return them all.

	for _, v := range opts.visitors {
		err := walk(v, root)
		if err != nil {
			if perr, ok := err.(*ParseError); ok {
				return nil, perr
			}
			return nil, genericParseError("validation failure: " + err.Error())
		}
		// Check to see if a MessageValidator had an error.
		// TODO: allow all to return.
		if mv, ok := v.(*MessageVisitor); ok {
			if mv.tape.firstErr != nil {
				return nil, mv.tape.firstErr.assErr()
			}
		}
	}

	// TODO: visitor to skip/count stars and warn here-ish.

	// opV := newopValidator()
	// ast.Walk(opV, root)
	// if opV.Err() != nil {
	// 	return nil, opV.Err()
	// }

	// warnW := newWarner(p.withWarnings)
	// ast.Walk(warnW, root)
	// if len(warnW.warnings) > 0 {
	// 	if !p.withWarnings {
	// 		return nil, warnW.warnings[0]
	// 	}
	// 	p.warnings = warnW.warnings
	// }
	return root, nil
}

func genericParseError(message string) *ParseError {
	return &ParseError{
		Position: ast.NoPosition(),
		Msg:      message,
	}
}

// ErrorWith allows using an AST node as the position basis for a parser error
// message.
func ErrorWith(node Positioned, message string) *ParseError {
	return &ParseError{
		Position: node.Pos(),
		Msg:      message,
	}
}

// ErrorAt allows returning a parser error with a manually-defined position.
// Useful for tweaking position from a node.
func ErrorAt(pos ast.Pos, message string) *ParseError {
	return &ParseError{
		Position: pos,
		Msg:      message,
	}
}

type ParserOpts struct {
	debug    bool
	visitors []Visitor
}

type Option func(*ParserOpts)

// Debug instructs the parser to print detailed parse information
func Debug() Option {
	return func(p *ParserOpts) {
		p.debug = true
	}
}

// Visitors adds additional visitors to the parsing pass
func Visitors(visitors ...Visitor) Option {
	return func(p *ParserOpts) {
		p.visitors = append(p.visitors, visitors...)
	}
}
