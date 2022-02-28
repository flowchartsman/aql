package lexer

import (
	"testing"

	"github.com/flowchartsman/aql/lexer/token"
)

func TestNextToken(t *testing.T) {
	input := `foo.bar[1]: "ba\"z" #comment
	#another comment
	OR 番号:<10gb
	and re:/^slash:\/\d$/
	OR before:within("12h")
	AND !(date: 1985-04-12T23:20:50.52Z or bob."my man":false)`
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "foo"},
		{token.DOT, "."},
		{token.IDENT, "bar"},
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.RBRACKET, "]"},
		{token.COLON, ":"},
		{token.STRING, `ba"z`},
		{token.OR, "OR"},
		{token.IDENT, "番号"},
		{token.COLON, ":"},
		{token.LT, "<"},
		{token.INT, "10gb"},
		{token.AND, "and"},
		{token.IDENT, "re"},
		{token.COLON, ":"},
		{token.REGEXP, `^slash:/\d$`},
		{token.OR, "OR"},
		{token.IDENT, "before"},
		{token.COLON, ":"},
		{token.IDENT, "within"},
		{token.LPAREN, "("},
		{token.STRING, "12h"},
		{token.RPAREN, ")"},
		{token.AND, "AND"},
		{token.NOT, "!"},
		{token.LPAREN, "("},
		{token.IDENT, "date"},
		{token.COLON, ":"},
		{token.TIMESTAMP, "1985-04-12T23:20:50.52Z"},
		{token.OR, "or"},
		{token.IDENT, "bob"},
		{token.DOT, "."},
		{token.STRING, "my man"},
		{token.COLON, ":"},
		{token.BOOL, "false"},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
