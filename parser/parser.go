package parser

//go:generate pigeon -no-recover -o parser-gen.go aql.peg

import (
	"fmt"
	"strings"
)

// NodeType represents the type of a node in the parse tree
type NodeType int

// Node Types
const (
	_ NodeType = iota
	NodeNot
	NodeAnd
	NodeOr
	NodeTerminal
)

// Node is a node in the query parse tree
type Node struct {
	NodeType   NodeType
	Comparison Comparison
	Left       *Node
	Right      *Node
}

// Comparison is an individual comparision operation on a terminal node
type Comparison struct {
	Op     string
	Field  []string
	Values []string
}

// ParseError is the exported error type for parsing errors with detailed information as to where they occurred
type ParseError struct {
	Inner    error    `json:"inner"`
	Line     int      `json:"line"`
	Column   int      `json:"column"`
	Offset   int      `json:"offset"`
	Prefix   string   `json:"prefix"`
	Expected []string `json:"expected"`
}

// Error Conforms to Error
func (p *ParseError) Error() string {
	return p.Prefix + ": " + p.Inner.Error()
}

func getRootNode(v interface{}) (*Node, error) {
	switch t := v.(type) {
	case nil:
		return nil, fmt.Errorf("parser returned nil output")
	case *Node:
		return t, nil
	default:
		return nil, fmt.Errorf("parser returned unknown type: %T", t)
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

// unused, may be useful for other matches (blatently stolen from mailgun article on similar language)
func toString(label interface{}) (string, error) {
	var sb strings.Builder
	value := label.([]interface{})
	for _, i := range value {
		if i == nil {
			continue
		}
		switch b := i.(type) {
		case []byte:
			sb.WriteByte(b[0])
		case string:
			sb.WriteString(b)
		case []interface{}:
			s, err := toString(i)
			if err != nil {
				return "", err
			}
			sb.WriteString(s)
		default:
			return "", fmt.Errorf("unexpected type [%T] found in label interfaces: %+v", i, i)
		}
	}
	return sb.String(), nil
}

// pigeon helper method, sometimes you gotta do what you gotta do
func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}

// helper method to get individual tokens from their rule index
func getTokens(first, rest interface{}, idx int) []string {
	out := []string{first.(string)}
	restSl := toIfaceSlice(rest)
	for _, v := range restSl {
		expr := toIfaceSlice(v)
		out = append(out, expr[idx].(string))
	}
	return out
}
