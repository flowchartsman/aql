package token

type Type string

type Token struct {
	Type    Type
	Literal string
	Offset  int
	Line    int
	Pos     int
	Err     string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers & literals
	IDENT     = "IDENT"     // path, func
	INT       = "INT"       // 42
	FLOAT     = "FLOAT"     // 1.02
	STRING    = "STRING"    // "hello"
	REGEXP    = "REGEXP"    // /(?i)^HI friends$/
	BOOL      = "BOOL"      // true, false
	TIMESTAMP = "TIMESTAMP" // 1985-04-12T23:20:50.52Z, 1985-04-12
	NET       = "NET"       // 192.168.10.0/24
	STAR      = "STAR"      // *

	// Delimiters
	DOT      = "."
	LPAREN   = "("
	RPAREN   = ")"
	LBRACKET = "["
	RBRACKET = "]"
	COMMA    = ","
	COLON    = ":"
	HASH     = "#"

	// Operators - logical
	AND  = "AND"
	OR   = "OR"
	NOT  = "NOT"
	BANG = "!"

	// Operators - comparison
	EQUALS = "=="
	NEQ    = "!="
	LT     = "<"
	LTE    = "<="
	GT     = ">"
	GTE    = ">="
	BET    = "><"
	SIM    = "~"

	// Operators - math
	PLUS  = `+`
	MINUS = `-`
	SLASH = `/`
)

func ClassifyAlphaLiteral(lit string) Type {
	return alphaLiteralClassifier.getLiteralType(lit)
}

var alphaLiteralClassifier = newClassifier(
	pattern{`(?i)true|false`, BOOL},
	pattern{`(?i)and`, AND},
	pattern{`(?i)or`, OR},
	pattern{`(?i)not`, NOT},
	pattern{`[\pL\pN_]+`, IDENT},
)

func ClassifyNumericLiteral(lit string) Type {
	return numericLiteralClassifier.getLiteralType(lit)
}

// add int multipliers like kb/mb etc
var numericLiteralClassifier = newClassifier(
	pattern{`-?\d+(?i:[kmg]b)?`, INT},
	pattern{`-?\d*\.\d+(?i:[kmg]b)?`, FLOAT},
	pattern{`(?:\d{1,3}\.){3}\d{1,3}(?:/\d{1,2})?`, NET},
	pattern{
		`\d{4}-\d{2}-\d{2}(?:[Tt ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:[Zz]|([+-]\d{2}:\d{2})))?`,
		TIMESTAMP,
	},
)
