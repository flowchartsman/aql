package parser

import "github.com/flowchartsman/aql/lexer/token"

const (
	_ int = iota
	LOWEST
	OR
	AND
	EQUALS     // ==
	COMPARISON // < <= > >= ~ and friends
	ADDSUB     // + -
	MULDIV     // * /
	PREFIX     // ! NOT -(negative)
	CALL       // function(X)
)

var precedences = map[token.Type]int{
	token.EQUALS: EQUALS,
	token.NEQ:    EQUALS,
	token.LT:     COMPARISON,
	token.LTE:    COMPARISON,
	token.GT:     COMPARISON,
	token.GTE:    COMPARISON,
	token.SIM:    COMPARISON,
	token.PLUS:   ADDSUB,
	token.MINUS:  ADDSUB,
	token.STAR:   MULDIV,
	token.SLASH:  MULDIV,
}
