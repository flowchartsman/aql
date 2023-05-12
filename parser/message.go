package parser

import (
	"encoding/json"
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
)

// MessageType is the type of a parser message
type MessageType int

func (t *MessageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(typeStr(*t))
}

const (
	msgMin MessageType = iota
	MsgHint
	MsgWarning
	MsgError
	msgMax
)

func typeStr(t MessageType) string {
	switch t {
	case MsgHint:
		return "HINT"
	case MsgWarning:
		return "WARNING"
	case MsgError:
		return "ERROR"
	default:
		return "<INVALID TYPE>"
	}
}

// Positioned refers to language features such as statements and identifiers
// that have a defined position in the query string.
// TODO: Consider just replacing with Pos in message and error calls for more
// precise control.
type Positioned interface {
	Pos() ast.Pos
}

// ParserMessage is a message output by the parser.
type ParserMessage struct {
	Position ast.Pos     `json:"position"`
	Msg      string      `json:"message"`
	Type     MessageType `json:"type"`
}

func (p *ParserMessage) String() string {
	if p.Position.Line == 0 || p.Position.Col == 0 || p.Position.Offset == -1 {
		return fmt.Sprintf("%s: %s", typeStr(p.Type), p.Msg)
	}
	return fmt.Sprintf("%s [%d:%d(%d)]: %s", typeStr(p.Type), p.Position.Line, p.Position.Col, p.Position.Offset, p.Msg)
}

func (p *ParserMessage) Message() string {
	return p.Msg
}

func (p *ParserMessage) Pos() ast.Pos {
	return p.Position
}

func (p *ParserMessage) assErr() *ParseError {
	return &ParseError{
		Position: p.Position,
		Msg:      p.Msg,
	}
}

func newMessage(t MessageType, pos ast.Pos, msg string) *ParserMessage {
	return &ParserMessage{
		Position: pos,
		Msg:      msg,
		Type:     t,
	}
}
