package parser

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseQuery(t *testing.T) {
	testParse(t,
		"simple minimal condition",
		`name:"siegfried"`,
		`(== name "siegfried")`)
	testParse(t,
		"quoted field",
		`"name":"siegfried"`,
		`(== name "siegfried")`)
	testParse(t,
		"field with leading underscore",
		`_name:"siegfried"`,
		`(== _name "siegfried")`)
	testParse(t,
		"field with leading underscore quoted",
		`"_name":"siegfried"`,
		`(== _name "siegfried")`)
	testParse(t,
		"field with just a number",
		`0:"siegfried"`,
		`(== 0 "siegfried")`)
	testParse(t,
		"multi part field name",
		`name.givenname:"siegfried"`,
		`(== name.givenname "siegfried")`)
	testParse(t,
		"quoted multi part field name",
		`name."GivenName":"siegfried"`,
		`(== name.GivenName "siegfried")`)
	testParse(t,
		"field name with dots and quotes",
		`"na.me"."Given\"Name":"siegfried"`,
		`(== "na.me"."Given\"Name" "siegfried")`)
	testParse(t,
		"simple AND clause",
		`name:"Hans" AND surname:"Wurst"`,
		`(&& (== name "Hans") (== surname "Wurst"))`)
	testParse(t,
		"simple AND clause with parenthesis",
		`(name:"Hans" AND surname:"Wurst")`,
		`(&& (== name "Hans") (== surname "Wurst"))`)
	testParse(t, "simple OR clause",
		`name:"Hans" OR name:"Siegfried"`,
		`(|| (== name "Hans") (== name "Siegfried"))`)
	testParse(t,
		"simple OR clause with parenthesis",
		`(name:"Hans" OR name:"Siegfried")`,
		`(|| (== name "Hans") (== name "Siegfried"))`)
	testParse(t,
		"simple OR clause with parenthesis around condition",
		`name:"Hans" OR (name:"Siegfried")`,
		`(|| (== name "Hans") (== name "Siegfried"))`)
	testParse(t,
		"simple AND clause with newline",
		"name:\"Hans\"\n\tAND surname:\"Wurst\"",
		`(&& (== name "Hans") (== surname "Wurst"))`)
	testParse(t,
		"simple OR clause with newline",
		"name:\"Hans\"\n\tOR surname:\"Wurst\"",
		`(|| (== name "Hans") (== surname "Wurst"))`)
	testParse(t,
		"OR / AND clauses",
		`name:"Hans" OR name:"Siegfried" AND age:9001`,
		`(|| (== name "Hans") (&& (== name "Siegfried") (== age 9001)))`)
	testParse(t,
		"OR / AND with paren precedence",
		`(name:"Hans" OR name:"Siegfried") AND age:9001`,
		`(&& (|| (== name "Hans") (== name "Siegfried")) (== age 9001))`)
	testParse(t,
		"OR / AND clauses reordered",
		`name:"Hans" AND age:9001 OR name:"Siegfried"`,
		`(|| (&& (== name "Hans") (== age 9001)) (== name "Siegfried"))`)
	testParse(t,
		"In syntax",
		`name:("Hans","Siegfried") AND age:9001`,
		`(&& (== name ["Hans", "Siegfried"]) (== age 9001))`)
	testParse(t, "simple NOT clause",
		`!name:"Hans" AND surname:"Wurst"`,
		`(&& (! (== name "Hans")) (== surname "Wurst"))`)
	testParse(t, "alternate simple NOT clause",
		`NOT name:"Hans" AND surname:"Wurst"`,
		`(&& (! (== name "Hans")) (== surname "Wurst"))`)
	testParse(t, "simple NOT clause with parenthesis",
		`!(name:"Hans" AND surname:"Wurst")`,
		`(! (&& (== name "Hans") (== surname "Wurst")))`)
	testParse(t, "float value",
		`floppy:1.4`,
		`(== floppy 1.4)`)
	testParse(t, "negative float value",
		`floppy:-1.4`,
		`(== floppy -1.4)`)
	testParse(t, "int value",
		`memory:32`,
		`(== memory 32)`)
	testParse(t, "negative int value",
		`memory:-32`,
		`(== memory -32)`)
	testParse(t, "boolean (true) value",
		`isAdmin:true`,
		`(== isAdmin true)`)
	testParse(t, "boolean (false) value",
		`writesGoodParsers:false`,
		`(== writesGoodParsers false)`)
	testParse(t,
		"net value",
		`internal:192.168.1.0/24`,
		`(== internal 192.168.1.0/24)`)
	testParse(t,
		"fullDate value",
		`Andy:1979-10-03`,
		`(== Andy 1979-10-03)`)
	testParse(t,
		"dateTime value",
		`AndyPrecise:2021-06-08T20:56:33+00:00`,
		`(== AndyPrecise 2021-06-08T20:56:33+00:00)`)
	testParse(t,
		"regexp value",
		`domains:/.*\\.[a-z0-9]*\\.local/`,
		`(== domains /.*\\.[a-z0-9]*\\.local/)`)
	testParse(t,
		"operator == (implicit)", `answer:42`,
		`(== answer 42)`)
	testParse(t,
		"operator ><",
		`whiskers:><(0,1)`,
		`(>< whiskers [0, 1])`)
	testParse(t,
		"operator >",
		`over9000:>9000`,
		`(> over9000 9000)`)
	testParse(t,
		"operator >=",
		`almost:>=9000`,
		`(>= almost 9000)`)
	testParse(t,
		"operator <",
		`alone:<2`,
		`(< alone 2)`)
	testParse(t,
		"operator <=",
		`pair:<=2`,
		`(<= pair 2)`)
	testParse(t,
		"operator exists",
		`pair:exists`,
		`(exists pair)`)
	testParse(t,
		"operator null",
		`pair:null`,
		`(null pair)`)
	testParse(t,
		"allow leading whitespace",
		` name:"Peter"`,
		`(== name "Peter")`)
	testParse(t,
		"allow trailing whitespace",
		`name:"Peter" `,
		`(== name "Peter")`)
	testParse(t,
		"allow whitespace",
		`name: ~ ( "Peter" , "Bob" )`,
		`(~ name ["Peter", "Bob"])`)
	testParse(t,
		"allow dash in field name",
		`na-me: "Peter"`,
		`(== na-me "Peter")`)
	testParse(t,
		"single parenthetical",
		`( name: "Peter" )`,
		`(== name "Peter")`)
	testParse(t,
		"mix of regular and  no-arg ops",
		`a:<1 AND b:exists AND c:<=2 AND d:null AND e:"hello"`,
		`(&& (< a 1) (&& (exists b) (&& (<= c 2) (&& (null d) (== e "hello")))))`)
	testParse(t,
		"subdoc node",
		`foo."ba r"{a:<1 AND b:"hello"}`,
		`(foo."ba r"{(&& (< a 1) (== b "hello"))})`)
}

func testParse(t *testing.T, testName string, query string, want string) {
	t.Helper()
	t.Run(testName, func(tt *testing.T) {
		tt.Helper()
		n, err := ParseQuery(query)
		if err != nil {
			tt.Fatalf("unexpected error: %v", err)
		}
		ns := n.String()
		if ns != want {
			tt.Fatalf("\nexpected:\n%s\ngot:\n%s", want, ns)
		}
	})
}

func TestParsingErrors(t *testing.T) {
	testParseErr(t,
		`unterminated string simple`,
		`name:"foo`,
		`1:10(9): unterminated string value, did you forget a closing '"'?`)
	// testParseErr(t,
	// 	`unterminated string middle`,
	// 	`foo:"foo AND bar:"bar"`,
	// 	` `)
}

func TestValueErrors(t *testing.T) {
	testParseErr(t,
		`invalid regexp fails`,
		`name:/*/`,
		"1:6(5): invalid regular expression [/*/]: error parsing regexp: missing argument to repetition operator: `*`")
	testParseErr(t,
		`valid regexp`,
		`name:/.*/`,
		``)
	testParseErr(t,
		`invalid net addr fails`,
		`net:500.500.500.500/32`,
		`1:5(4): invalid network value [500.500.500.500/32]: IPv4 field has value >255`)
	testParseErr(t,
		`invalid net block fails`,
		`net:192.168.0.0/99`,
		`1:5(4): invalid network value [192.168.0.0/99]: prefix length out of range`)
	testParseErr(t,
		`valid net block`,
		`net:192.168.0.0/24`,
		``)
	testParseErr(t,
		`invalid short date month fails`,
		`Andy:1979-13-03`,
		`1:6(5): invalid datetime value [1979-13-03]: month out of range`)
	testParseErr(t,
		`invalid short date day fails`,
		`Joe:1979-02-31`,
		`1:5(4): invalid datetime value [1979-02-31]: day out of range`)
	testParseErr(t,
		`valid short date`,
		`Andy:1979-10-03`,
		``)
	testParseErr(t,
		`invalid long date time fails`,
		`AndyPrecise:2021-06-08T20:74:33+00:00`,
		`1:13(12): invalid datetime value [2021-06-08T20:74:33+00:00]: minute out of range`)
	testParseErr(t,
		`valid long date`,
		`AndyPrecise:2021-06-08T20:53:33+00:00`,
		``)
	testParseErr(t,
		`invalid value in list fails`,
		`name:(/.*/,/*/)`,
		"1:12(11): invalid regular expression [/*/]: error parsing regexp: missing argument to repetition operator: `*`")
	testParseErr(t,
		`unnecessary paren fails`,
		`foo:("bar")`,
		"1:5(4): unnecessary parenthesis for only one value")
	testParseErr(t,
		`empty paren fails`,
		`name:()`,
		"1:7(6): unexpected closing parenthesis, expecting values")
}

func TestOpErrors(t *testing.T) {
	testParseErr(t,
		`duplicates not allowed`,
		`value: (1,2,1)`,
		`1:13(12): duplicate argument [1] (value 3/3)`)
	testParseErr(t,
		`between operator invalid arity 1`,
		`value:>< 1`,
		`1:1(0): [><] operation requires exactly 2 arguments`)
	testParseErr(t,
		`between operator invalid arity >2`,
		`value:>< (1,2,3)`,
		`1:1(0): [><] operation requires exactly 2 arguments`)
	testParseErr(t,
		`between operator needs two numeric arguments`,
		`value:>< (1, "hello")`,
		`1:14(13): [><] operation needs numeric arguments`)
	testParseErr(t,
		`between operator requires second value to be greater`,
		`value:>< (2, 1)`,
		`1:14(13): [><] operation requires the second argument be greater`)
	// ensure numeric requirements
	for _, op := range []string{`<`, `<=`, `>`, `>=`, `><`} {
		query := fmt.Sprintf(`value:%s "hello"`, op)
		expectedErr := fmt.Sprintf(`*[%s] operation needs numeric arguments`, op)
		testName := fmt.Sprintf(`operation %s requires numeric value(s)`, op)
		testParseErr(t,
			testName,
			query,
			expectedErr,
		)
	}
	testParseErr(t,
		`similarity operator needs string values`,
		`value:~ 2`,
		`1:9(8): [~] operation needs string arguments`)
}

func testParseErr(t *testing.T, testName string, query string, wantErr string) {
	t.Helper()
	t.Run(testName, func(t *testing.T) {
		t.Helper()
		_, err := ParseQuery(query)
		if err != nil {
			t.Log(PrettyErr(query, err))
			if wantErr == "" {
				t.Fatalf("\nunexpected error:\n%s", err)
			}
			tv := false
			if wantErr[0] == '*' {
				tv = strings.HasSuffix(err.Error(), wantErr[1:])
			} else {
				tv = err.Error() == wantErr
			}
			if !tv {
				t.Fatalf("\nexpected:\n%s\ngot:\n%s", wantErr, err)
			}
		} else {
			if wantErr != "" {
				t.Fatalf("\nexpected:\n%s\ngot:\n(no error)", wantErr)
			}
		}
	})
}

// TODO: full coverage tests
// func TestCoverage(t *testing.T){/**/}

// testPanic(t *testing.T, expectedPanicMsg string, do func()){/**/}
