package parser

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flowchartsman/aql/ast"
	"github.com/flowchartsman/aql/lexer"
	"github.com/flowchartsman/aql/lexer/token"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type ParserError struct {
	msg     string
	offset  int
	line    int
	linepos int
}

func (pe *ParserError) String() string {
	return fmt.Sprintf("[%d:%d] %s", pe.line, pe.linepos, pe.msg)
}

type Parser struct {
	l              *lexer.Lexer
	errors         []ParserError
	curToken       token.Token
	peekToken      token.Token
	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l: l,
		// can also just add the references here, directly...
		prefixParseFns: map[token.Type]prefixParseFn{},
		infixParseFns:  map[token.Type]infixParseFn{},
	}
	// hydrate curToken/peekToken
	p.nextToken()
	p.nextToken()
	// register TDOP functions
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.STRING, p.parseIdentifier)

	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.REGEXP, p.parseRegexpLiteral)
	p.registerPrefix(token.NET, p.parseNetLiteral)
	p.registerPrefix(token.TIMESTAMP, p.parseTimestampLiteral)

	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)

	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.STAR, p.parseInfixExpression)
	p.registerInfix(token.EQUALS, p.parseInfixExpression)
	p.registerInfix(token.NEQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []ParserError {
	return p.errors
}

func (p *Parser) setParseError(failMsg string, badToken token.Token, wantedTypes ...token.Type) {
	var errMsg strings.Builder
	errMsg.WriteString(failMsg)
	errMsg.WriteString(` -- found: `)
	if badToken.Type == token.ILLEGAL && badToken.Err != "" {
		errMsg.WriteString(`<`)
		errMsg.WriteString(badToken.Err)
		errMsg.WriteString(`>`)
	} else {
		errMsg.WriteString(fmt.Sprintf("%q", badToken.Literal))
		if string(badToken.Type) != badToken.Literal {
			errMsg.WriteString(fmt.Sprintf("(%s)", badToken.Type))
		}
	}
	if len(wantedTypes) > 0 {
		errMsg.WriteString(" wanted: ")
		for i, wt := range wantedTypes {
			errMsg.WriteString(string(wt))
			if i < len(wantedTypes)-1 {
				errMsg.WriteString("|")
			}
		}
	}
	p.errors = append(p.errors, ParserError{
		msg:     errMsg.String(),
		offset:  badToken.Offset,
		line:    badToken.Line,
		linepos: badToken.Pos,
	})
}

func (p *Parser) registerPrefix(tType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tType] = fn
}

func (p *Parser) registerInfix(tType token.Type, fn infixParseFn) {
	p.infixParseFns[tType] = fn
}

func (p *Parser) noPrefixParseFnError(t token.Token) {
	p.setParseError(fmt.Sprintf("no prefix function found for token type: %s", t.Type), t)
}

func (p *Parser) noInfixParseFnError(t token.Token) {
	p.setParseError(fmt.Sprintf("no infix function found for token type: %s", t.Type), t)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// XXX: do we really need query? maybe expression is at top, since this all evaluates to true/false at the end, and each piece is just a subexpression
func (p *Parser) ParseQuery() *ast.Query {
	query := &ast.Query{}
	if p.curToken.Type == token.EOF {
		return nil
	}
	query.Root = p.parseExpression(LOWEST)
	if query.Root == nil {
		return nil
	}
	p.nextToken()
	if p.peekToken.Type != token.EOF {
		p.setParseError("extra token at query end", p.peekToken, token.EOF)
		return nil
	}
	return query
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken)
		return nil
	}
	leftExp := prefix()

	for p.peekToken.Type != (token.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	field := p.parseField()
	if field == nil {
		return nil
	}
	return &ast.Identifier{
		Field: field,
	}
}

func (p *Parser) parseField() *ast.Field {
	switch p.curToken.Type {
	case token.IDENT, token.STRING:
		//good
	default:
		p.setParseError("field needs to begin with string or identifier", p.curToken)
		return nil
	}
	f := &ast.Field{}
	//append the first identifier
	f.PathSegments = append(f.PathSegments, ast.StringPathSegment(p.curToken.Literal))
	for {
	NEXT_SEGMENT:
		switch p.peekToken.Type {
		case token.DOT:
			goto CONSUME_SEGMENT
		case token.LBRACKET:
			goto CONSUME_INDEX
		default:
			return f
		}
	CONSUME_SEGMENT:
		p.nextToken()
		f.Tokens = append(f.Tokens, p.curToken)
		p.nextToken()
		switch p.curToken.Type {
		case token.STAR, token.IDENT, token.STRING:
			f.PathSegments = append(f.PathSegments, ast.StringPathSegment(p.curToken.Literal))
		default:
			p.setParseError("was looking for a field piece", p.curToken, token.IDENT, token.STRING)
			return nil
		}
		f.Tokens = append(f.Tokens, p.curToken)
		goto NEXT_SEGMENT
	CONSUME_INDEX:
		p.nextToken()
		f.Tokens = append(f.Tokens, p.curToken)
		p.nextToken()
		if p.curToken.Type != token.INT {
			p.setParseError("invalid index", p.curToken, token.INT)
			return nil
		}
		f.Tokens = append(f.Tokens, p.curToken)
		idx, _ := strconv.Atoi(p.curToken.Literal)
		f.PathSegments = append(f.PathSegments, ast.IndexPathSegment(idx))
		p.nextToken()
		if p.curToken.Type != token.RBRACKET {
			p.setParseError("missing index terminator", p.curToken, token.LBRACKET)
			return nil
		}
		f.Tokens = append(f.Tokens, p.curToken)
		goto NEXT_SEGMENT
	}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.setParseError("could not parse token as integer", p.curToken)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.setParseError("could not parse token as float64", p.curToken)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseRegexpLiteral() ast.Expression {
	lit := &ast.RegexpLiteral{Token: p.curToken}
	value, err := regexp.Compile(p.curToken.Literal)
	if err != nil {
		p.setParseError("could not parse token as regular expression", p.curToken)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseTimestampLiteral() ast.Expression {
	lit := &ast.TimestampLiteral{Token: p.curToken}
	format := time.RFC3339
	if len(p.curToken.Literal) == 10 {
		format = "2006-01-02"
	}
	value, err := time.Parse(format, p.curToken.Literal)
	if err != nil {
		p.setParseError("could not parse token a timestamp", p.curToken)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseNetLiteral() ast.Expression {
	lit := &ast.NetLiteral{Token: p.curToken}
	_, value, err := net.ParseCIDR(p.curToken.Literal)
	if err != nil {
		p.setParseError("could not parse token as net address", p.curToken)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExp{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}
