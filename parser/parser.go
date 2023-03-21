package parser

//go:generate pigeon -o parser-gen.go aql.peg

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

// ParseWarning is an warning error message that can be made non-fatal with the
// [WithWarnings] option.
type ParseWarning = grammar.ParseError

type Parser struct {
	debug        bool
	initGoTypes  bool
	withWarnings bool
	warnings     []*ParseWarning
}

func (p *Parser) ParseQuery(query string) (ast.Node, error) {
	return p.ParseQueryReader(strings.NewReader(query))
}

func (p *Parser) ParseQueryReader(r io.Reader) (ast.Node, error) {
	v, err := grammar.ParseReader("", r, grammar.Debug(p.debug))
	if err != nil {
		return nil, grammar.GetParseError(err)
	}
	// get number of stars skipped and warn here.
	root, err := getRootNode(v)
	if err != nil {
		return nil, err
	}
	if p.initGoTypes {
		i := newInitializer()
		ast.Walk(i, root)
		if i.Err() != nil {
			return nil, i.Err()
		}
	}
	opV := newopValidator()
	ast.Walk(opV, root)
	if opV.Err() != nil {
		return nil, opV.Err()
	}

	warnW := newWarner(p.withWarnings)
	ast.Walk(warnW, root)
	if len(warnW.warnings) > 0 {
		if !p.withWarnings {
			return nil, warnW.warnings[0]
		}
		p.warnings = warnW.warnings
	}
	return root, nil
}

func NewParser(opts ...Option) *Parser {
	p := &Parser{}
	for _, o := range opts {
		o(p)
	}
	return p
}

type Option func(*Parser) error

// Debug instructs the parser to print detailed parse information
func Debug() Option {
	return func(p *Parser) error {
		p.debug = true
		return nil
	}
}

// InitGoTypes instructs the parser to initialize the underlying Go types for
// all ast.Val structs, checking to ensure that they are valid. This reports,
// for example, invalid Go regular expressions, timestamps or CIDR netblocks
func InitGoTypes() Option {
	return func(p *Parser) error {
		p.initGoTypes = true
		return nil
	}
}

// WithWarnings instrucs the parser to emit warnings separately that can be
// viewed with the [parser.Warnings] call. Otherwise, warnings will be returned
// as errors that will terminate the parser.
func WithWarnings() Option {
	return func(p *Parser) error {
		p.withWarnings = true
		return nil
	}
}

// GetPrintableError returns a string suitable for printing to a terminal, replete with a handy caret indicator
func GetPrintableError(query string, err error) string {
	pe, ok := err.(*ParseError)
	if !ok {
		return err.Error()
	}
	var sb strings.Builder
	lines := strings.Split(query, "\n")
	badLine := lines[pe.Line-1]
	sb.WriteString(err.Error() + "\n")
	sb.WriteString(badLine + "\n")
	if pe.Column > 0 {
		sb.WriteString(strings.Repeat(`~`, pe.Column-1))
	}
	sb.WriteString("^\n")
	return sb.String()
}

func getRootNode(v interface{}) (ast.Node, error) {
	switch t := v.(type) {
	case nil:
		return nil, fmt.Errorf("parser returned nil output")
	case ast.Node:
		return t, nil
	default:
		return nil, fmt.Errorf("parser returned unknown type: %T", t)
	}
}
