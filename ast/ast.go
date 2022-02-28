package ast

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flowchartsman/aql/lexer/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

// might not need node
type Expression interface {
	Node
	expressionNode()
}

// put this everywhere a literal is used
type Literal interface {
	Expression
	literalNode()
}

type Query struct {
	Root Expression
}

func (q *Query) TokenLiteral() string {
	if q.Root != nil {
		return q.Root.TokenLiteral()
	}
	return ""
}

func (q *Query) String() string {
	if q.Root != nil {
		return q.Root.String()
	}
	return ""
}

type PathSegment struct {
	key string
	idx int
}

func StringPathSegment(s string) PathSegment {
	return PathSegment{
		key: s,
		idx: -1,
	}
}

func IndexPathSegment(i int) PathSegment {
	return PathSegment{
		key: "",
		idx: i,
	}
}

type Field struct {
	PathSegments []PathSegment
	Tokens       []token.Token
	// soon
	// Regexp *regexp.Regexp
}

func (f *Field) String() string {
	var sb strings.Builder
	for i, s := range f.PathSegments {
		switch {
		case s.idx >= 0:
			sb.WriteString(strconv.Itoa(s.idx))
		case s.key != "":
			sb.WriteString(strconv.Quote(s.key))
		}
		if i < len(f.PathSegments)-1 {
			sb.WriteString(`,`)
		}
	}
	return sb.String()
}

type Identifier struct {
	Field *Field
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	var sb strings.Builder
	for _, t := range i.Field.Tokens {
		sb.WriteString(t.Literal)
	}
	return sb.String()
}
func (i *Identifier) String() string { return i.Field.String() }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type RegexpLiteral struct {
	Token token.Token
	Value *regexp.Regexp
}

func (rl *RegexpLiteral) expressionNode()      {}
func (rl *RegexpLiteral) TokenLiteral() string { return rl.Token.Literal }
func (rl *RegexpLiteral) String() string       { return rl.Token.Literal }

type TimestampLiteral struct {
	Token token.Token
	Value time.Time
}

func (tl *TimestampLiteral) expressionNode()      {}
func (tl *TimestampLiteral) TokenLiteral() string { return tl.Token.Literal }
func (tl *TimestampLiteral) String() string       { return tl.Token.Literal }

type NetLiteral struct {
	Token token.Token
	Value *net.IPNet
}

func (nl *NetLiteral) expressionNode()      {}
func (nl *NetLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NetLiteral) String() string       { return nl.Token.Literal }

type PrefixExp struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExp) expressionNode() {}
func (pe *PrefixExp) TokenLiteral() string {
	return pe.Token.Literal
}
func (pe *PrefixExp) String() string {
	var sb strings.Builder
	sb.WriteString(`(`)
	sb.WriteString(pe.Operator)
	sb.WriteString(pe.Right.String())
	sb.WriteString(`)`)
	return sb.String()
}

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}
func (ie *InfixExpression) TokenLiteral() string {
	return ie.Token.Literal
}
func (ie *InfixExpression) String() string {
	var sb strings.Builder
	sb.WriteString(`(`)
	sb.WriteString(ie.Left.String())
	sb.WriteString(` ` + ie.Operator + ` `)
	sb.WriteString(ie.Right.String())
	sb.WriteString(`)`)
	return sb.String()
}
