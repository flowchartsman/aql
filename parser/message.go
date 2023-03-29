package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	msgMax
)

func typeStr(t MessageType) string {
	switch t {
	case MsgHint:
		return "HINT"
	case MsgWarning:
		return "WARNING"
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

func newMessage(t MessageType, pos ast.Pos, msg string) *ParserMessage {
	return &ParserMessage{
		Position: pos,
		Msg:      msg,
		Type:     t,
	}
}

// PrettyPrintable represents messages that can be pretty-printed
type PrettyPrintable interface {
	Pos() ast.Pos
	Message() string
}

// PrettyMessages returns string representations of parser messages that are
// suitable for printing to a terminal, along with a handy caret indicator
// pointing to the relevant position in the query, if applicable. Because this
// function only splits up the query once, it should be preferred when
// pretty-printing multiple messages.
func PrettyMessages(query string, messages ...*ParserMessage) []string {
	if len(messages) == 0 {
		return nil
	}
	queryLines := strings.Split(query, "\n")
	out := make([]string, 0, len(messages))
	for _, m := range messages {
		out = append(out, prettyPrint(queryLines, m))
	}
	return out
}

// PrettyMessage is a convenience method for pretty printing a single parser
// message. Since the query will be split on each call, consider using
// [PrettyMessages] if you have a larger query
func PrettyMessage(query string, message *ParserMessage) string {
	queryLines := strings.Split(query, "\n")
	return prettyPrint(queryLines, message)
}

// PrettyErr returns the pretty representation of a parser error message, if
// it contains location information. Othwrwiese it will just print the error.
func PrettyErr(query string, err error) string {
	var pe *ParseError
	if errors.As(err, &pe) {
		queryLines := strings.Split(query, "\n")
		return prettyPrint(queryLines, pe)
	}
	return err.Error()
}

func prettyPrint(queryLines []string, message PrettyPrintable) string {
	if message.Pos().IsZero() {
		message.Message()
	}
	// TODO: trim lines above a certain threshold here.
	var sb strings.Builder
	sb.WriteString(message.Message())
	sb.WriteString("\n")
	sb.WriteString(queryLines[message.Pos().Line-1])
	sb.WriteString("\n")
	if message.Pos().Col > 0 {
		sb.WriteString(strings.Repeat(` `, message.Pos().Col-1))
	}
	sb.WriteString(strings.Repeat(`^`, message.Pos().Len))
	return sb.String()
}
