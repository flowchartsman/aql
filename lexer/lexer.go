package lexer

import (
	"strings"
	"unicode"

	"github.com/flowchartsman/aql/lexer/token"
)

const EOF rune = '\u0000'

// Lexer is the AQL lexer
type Lexer struct {
	input         []rune
	position      int
	readPosition  int
	ch            rune
	line          int
	linePos       int
	previousToken token.Token
}

func New(input string) *Lexer {
	l := &Lexer{
		// TODO: []rune(string) is convenient, but expensive, better to refactor parser to readrune, which can then conceivably buffer better
		input: []rune(input),
		line:  1,
		// empty token to start
		previousToken: token.Token{
			Type: "",
		},
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = EOF
		// TODO: can return early here?
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.linePos += 1
	l.readPosition += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	defer func() {
		l.previousToken = tok
	}()

	for {
		l.skipWhitespace()
		if l.ch == '#' {
			l.skipComment()
			continue
		}
		break
	}

	switch l.ch {
	case '.':
		tok = l.emit(token.DOT, string(l.ch))
	case '(':
		tok = l.emit(token.LPAREN, string(l.ch))
	case ')':
		tok = l.emit(token.RPAREN, string(l.ch))
	case '[':
		tok = l.emit(token.LBRACKET, string(l.ch))
	case ']':
		tok = l.emit(token.RBRACKET, string(l.ch))
	case ',':
		tok = l.emit(token.COMMA, string(l.ch))
	case ':':
		tok = l.emit(token.COLON, string(l.ch))
	case '~':
		tok = l.emit(token.SIM, string(l.ch))
	case '!':
		if l.peek() == '=' {
			tok = l.emit(token.NEQ, "!=")
			l.readChar()
		}
		tok = l.emit(token.BANG, string(l.ch))
	case '+':
		tok = l.emit(token.PLUS, string(l.ch))
	case '-':
		tok = l.emit(token.MINUS, string(l.ch))
	case '=':
		if l.peek() == '=' {
			tok = l.emit(token.EQUALS, "==")
			l.readChar()
		}
	case '<':
		if l.peek() == '=' {
			tok = l.emit(token.LTE, "<=")
			l.readChar()
		} else {
			tok = l.emit(token.LT, "<")
		}
	case '>':
		switch l.peek() {
		case '=':
			tok = l.emit(token.GTE, ">=")
			l.readChar()
		case '<':
			tok = l.emit(token.BET, "><")
			l.readChar()
		default:
			tok = l.emit(token.GT, ">")
		}
	case '"':
		tok = l.readEnclosedLiteral('"', token.STRING)
	case '/':
		switch l.previousToken.Type {
		case token.STRING, token.REGEXP, token.IDENT, token.INT, token.FLOAT, token.TIMESTAMP, token.NET:
			tok = l.emit(token.SLASH, "/")
		default:
			tok = l.readEnclosedLiteral('/', token.REGEXP)
		}
	case EOF:
		tok = token.Token{
			Type:    token.EOF,
			Literal: "",
		}
	default:
		switch {
		case unicode.IsLetter(l.ch):
			// return early, eats chars. TODO: Fix eating to clear up last token logic and remove defer
			tok = l.readAlphaLiteral()
			return tok
		case unicode.IsDigit(l.ch):
			// return early, eats chars
			tok = l.readNumericLiteral()
			return tok
		default:
			tok = l.emit(token.ILLEGAL, string(l.ch))
		}
	}
	l.readChar()
	return tok
}

func (l *Lexer) peek() rune {
	if l.readPosition >= len(l.input) {
		return EOF
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		if l.ch == '\n' {
			l.line++
			l.linePos = 1
		}
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	l.readChar()
	for !(l.ch == '\n' || l.ch == EOF) {
		l.readChar()
	}
	if l.ch == '\n' {
		l.line++
		l.linePos = 1
	}
}

func (l *Lexer) emit(tokenType token.Type, literal string) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: literal,
		Offset:  l.position,
		// todo: consider readpos as well for highlighting
		Line: l.line,
		Pos:  l.linePos,
	}
}

func (l *Lexer) readAlphaLiteral() token.Token {
	position := l.position
	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) {
		l.readChar()
	}
	lit := string(l.input[position:l.position])
	ttype := token.ClassifyAlphaLiteral(lit)
	return token.Token{
		Type:    ttype,
		Literal: lit,
		Offset:  l.position - len(lit),
		Line:    l.line,
		Pos:     l.linePos - len(lit),
	}
}

func (l *Lexer) readNumericLiteral() token.Token {
	position := l.position
	for !isDelimit(l.ch) {
		l.readChar()
	}
	lit := string(l.input[position:l.position])
	ttype := token.ClassifyNumericLiteral(lit)
	return token.Token{
		Type:    ttype,
		Literal: lit,
		Offset:  l.position - len(lit),
		Line:    l.line,
		Pos:     l.linePos - len(lit),
	}
}

func (l *Lexer) readEnclosedLiteral(marker rune, want token.Type) token.Token {
	position := l.position + 1
	for l.ch != EOF {
		l.readChar()
		if l.ch == '\\' && l.peek() == marker {
			l.readChar()
			l.readChar()
		}
		if l.ch == marker {
			break
		}
	}
	lit := string(l.input[position:l.position])
	lit = strings.Replace(lit, `\`+string(marker), string(marker), -1)

	tok := token.Token{
		Type:    want,
		Literal: lit,
		Offset:  l.position - (len(lit) + 1),
		Line:    l.line,
		Pos:     l.linePos - (len(lit) + 1),
		Err:     "",
	}

	if l.ch == EOF {
		tok.Type = token.ILLEGAL
		tok.Err = "unterminated " + string(want) + " literal"
	}
	return tok
}

// necessary to keep numeric literals from getting too big for their britches
// specifically for operands of the form a.b[1].
// TODO: return to this for context-sensitive lexing
func isDelimit(r rune) bool {
	if unicode.IsSpace(r) {
		return true
	}
	switch r {
	case '[', ']', '(', ')', ',', EOF:
		return true
	default:
		return false
	}
}
