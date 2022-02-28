package parser

import (
	"reflect"
	"testing"

	"github.com/flowchartsman/aql/ast"
	"github.com/flowchartsman/aql/lexer"
)

func TestIdentifierExpression(t *testing.T) {
	var typeRef *ast.Identifier
	ensureParserOut(t, `field`, typeRef, `"field"`)
	ensureParserOut(t, `field.foo`, typeRef, `"field","foo"`)
	ensureParserOut(t, `field.foo[0]`, typeRef, `"field","foo",0`)
	ensureParserOut(t, `field.foo[0].bar`, typeRef, `"field","foo",0,"bar"`)
	ensureParserOut(t, `field.foo."0"`, typeRef, `"field","foo","0"`)
	ensureParserOut(t, `field.foo[0][1]`, typeRef, `"field","foo",0,1`)
	ensureParserOut(t, `field."hello world".foo`, typeRef, `"field","hello world","foo"`)
	ensureParserOut(t, `field.番.foo.号`, typeRef, `"field","番","foo","号"`)

	ensureErr(t, `field..foo`, `[1:7] was looking for a field piece -- found: "." wanted: IDENT|STRING`)
	ensureErr(t, `field."hello`, `[1:7] was looking for a field piece -- found: <unterminated STRING literal> wanted: IDENT|STRING`)
	ensureErr(t, `field./hello`, `[1:7] was looking for a field piece -- found: <unterminated REGEXP literal> wanted: IDENT|STRING`)
}

func TestLiteralExpressions(t *testing.T) {
	ensureParserOut(t, `1`, new(ast.IntegerLiteral), `1`)
	ensureParserOut(t, `1.0`, new(ast.FloatLiteral), `1.0`)
	ensureParserOut(t, `/(?i)^hello$/`, new(ast.RegexpLiteral), `(?i)^hello$`)
	ensureParserOut(t, `1979-01-01`, new(ast.TimestampLiteral), `1979-01-01`)
}

func TestPrefixExpression(t *testing.T) {
	var typeRef *ast.PrefixExp
	ensureParserOut(t, `!foo.bar`, typeRef, `(!"foo","bar")`)
	ensureParserOut(t, `NOT foo.bar`, typeRef, `(NOT"foo","bar")`)
	ensureParserOut(t, `-1`, typeRef, `(-1)`)
	// this should definitely fail for a better reason
	//ensureParserOut(t, `!f!o.bar`, typeRef, `(!"foo","bar")`)
}

func TestInfixExpression(t *testing.T) {
	var typeRef *ast.InfixExpression
	ensureParserOut(t, `foo.bar + 5`, typeRef, `(!"foo","bar")`)
	//ensureParserOut(t, `NOT foo.bar`, typeRef, `(NOT"foo","bar")`)
	// this should definitely fail for a better reason
	//ensureParserOut(t, `!f!o.bar`, typeRef, `(!"foo","bar")`)
}

func TestOperatorPrecedence(t *testing.T) {
	var typeRef *ast.InfixExpression
	ensureParserOut(t, `foo.bar + 5 / bar.foo`, typeRef, `(!"foo","bar")`)
}

func getParserOut(t *testing.T, input string) (*ast.Query, []ParserError) {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	q := p.ParseQuery()
	return q, p.Errors()
}

func ensureParserOut(t *testing.T, input string, wantType interface{}, wantOut string) {
	t.Helper()
	switch {
	case input == "":
		panic("ensureParserOut called with empty input")
	case wantType == nil:
		panic("ensureParserOut called with nil wantType")
	case wantOut == "":
		panic("ensureParserOut called with empty wantOut")
	}
	q, errors := getParserOut(t, input)
	if len(errors) != 0 {
		t.Logf("found %d unexpected errors:", len(errors))
		for _, e := range errors {
			t.Logf("\t%s", e.String())
		}
		t.FailNow()
	}
	if reflect.TypeOf(wantType) != reflect.TypeOf(q.Root) {
		t.Fatalf("type mismatch, wanted: %T got %T", wantType, q.Root)
	}
	if wantOut != q.String() {
		t.Fatalf("wanted: %s\n got:%s", wantOut, q.String())
	}
}

func ensureErr(t *testing.T, input string, wantErr string) {
	t.Helper()
	switch {
	case input == "":
		panic("ensureErr called with empty input")
	case wantErr == "":
		panic("ensureErr called with empty wantErr")
	}
	q, errors := getParserOut(t, input)
	if q != nil {
		t.Fatalf("query was not nil")
	}
	if len(errors) == 0 {
		t.Fatalf("wanted an error, but got none")
	}
	for _, e := range errors {
		if wantErr == e.String() {
			return
		}
	}
	t.Logf("found %d errors, but wanted error not found: %q", len(errors), wantErr)
	for _, e := range errors {
		t.Logf("\t%s", e.String())
	}
	t.FailNow()
}
