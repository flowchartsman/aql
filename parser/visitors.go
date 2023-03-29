package parser

import (
	"errors"
	"fmt"

	"github.com/flowchartsman/aql/parser/ast"
)

var (
	Skip    = errors.New("skipping node")
	SkipAll = errors.New("skipping all remaining nodes")
)

// Visitor is the interface that defines an AST visitor, which will visit nodes
// in a depth-first order. If it returns an error, crawling will be stopped
// immediately. *[ParseError]s will be treated specially and printed with the
// positional information. Optionally, it can also return [Skip] to skip
// visiting a particular node or [SkipAll] to stop processing.
type Visitor interface {
	Visit(node ast.Node) error
}

func walk(v Visitor, node ast.Node) error {
	err := walkr(v, node)
	if err == Skip || err == SkipAll {
		return nil
	}
	return err
}

// this can be simplified by giving nodes children
func walkr(v Visitor, node ast.Node) error {
	err := v.Visit(node)
	if err != nil {
		if err == Skip {
			return nil
		}
		return err
	}
	switch n := node.(type) {
	case *ast.AndNode:
		err = walkr(v, n.Left)
		if err != Skip {
			return err
		}
		return walkr(v, n.Right)
	case *ast.OrNode:
		err = walkr(v, n.Left)
		if err != Skip {
			return err
		}
		return walkr(v, n.Right)
	case *ast.NotNode:
		return walkr(v, n.Expr)
	case *ast.SubdocNode:
		return walkr(v, n.Expr)
	}
	return nil
}

// VisitorFun is a function that can act as a visitor, meaning it will be called
// for every node in the tree, unless it chooses to skip. Non-skip errors will
// halt the walk immediately.
type VisitorFunc func(node ast.Node) error

func (vf VisitorFunc) Visit(node ast.Node) error {
	return vf(node)
}

// MessageVisitor is a Visitor to generate informational messages.
type MessageVisitor struct {
	f    func(ast.Node, *MessageTape) error
	tape *MessageTape
}

// Visit implements [Visitor]
func (mv *MessageVisitor) Visit(node ast.Node) error {
	return mv.f(node, mv.tape)
}

// Messages can be used to retrieve the messages from a MessageVisitor when the
// walk is complete.
func (mv *MessageVisitor) Messages() []*ParserMessage {
	return mv.tape.messages
}

// NewMessageVisitor creates a visitor that can return [ParserMessage]s for
// errors, warnings and hints. The function will have access to a *[MessageTape]
// type to which messages can be appended. After parsing is complete, the
// messages can be retrieved with the [MessageVisitor.Messages] method.
func NewMessageVisitor(f func(ast.Node, *MessageTape) error) *MessageVisitor {
	return &MessageVisitor{
		f:    f,
		tape: &MessageTape{},
	}
}

// MessageTape is a message appended to be used during tree-traversal.
type MessageTape struct {
	messages []*ParserMessage
}

// Hint adds an informational message to the tape that is neither a warning nor
// an error, yet might be helpful to the user. If a particular query or regular
// expression is less efficient, this is the way to let the user know. Messages
// generated with this call will not have a position attached.
func (mt *MessageTape) Hint(msg string, v ...any) {
	mt.addMsg(MsgHint, ast.NoPosition(), msg, v...)
}

// HintAt adds a hint message to the tape with a query offset attached for
// printing or highlighting.
func (mt *MessageTape) HintAt(where ast.Pos, msg string, v ...any) {
	mt.addMsg(MsgHint, where, msg, v...)
}

// Warning adds a message to the tape that is a more notable, yet still not an
// error. Messages generated with this call will not have a positition attached.
func (mt *MessageTape) Warning(msg string, v ...any) {
	mt.addMsg(MsgWarning, ast.NoPosition(), msg, v...)
}

// WarningAt adds a warning message to the tape with a query offset attached for printing.
func (mt *MessageTape) WarningAt(where ast.Pos, msg string, v ...any) {
	mt.addMsg(MsgWarning, where, msg, v...)
}

func (mt *MessageTape) addMsg(msgType MessageType, where ast.Pos, msg string, v ...any) {
	mt.messages = append(mt.messages, newMessage(msgType, where, fmt.Sprintf(msg, v...)))
}
