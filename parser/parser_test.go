package parser

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParseQuery(t *testing.T) {
	t.Run("simple minimal condition", testParseQuery(`name:"siegfried"`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`name`},
			Values: []string{`siegfried`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("quoted field", testParseQuery(`"name":"siegfried"`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`name`},
			Values: []string{`siegfried`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("field with leading underscore", testParseQuery(`_name:"siegfried"`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`_name`},
			Values: []string{`siegfried`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("field with leading underscore qupted", testParseQuery(`"_name":"siegfried"`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`_name`},
			Values: []string{`siegfried`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("field with just a number", testParseQuery(`0:"siegfried"`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`0`},
			Values: []string{`siegfried`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("multi part field name", testParseQuery(`name.givenname:"siegfried"`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`name`, `givenname`},
			Values: []string{`siegfried`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("quoted multi part field name", testParseQuery(`name."GivenName":"siegfried"`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`name`, `GivenName`},
			Values: []string{`siegfried`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("field name with dots and quotes", testParseQuery(`"na.me"."Given\"Name":"siegfried"`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`na.me`, `Given"Name`},
			Values: []string{"siegfried"},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("simple AND clause", testParseQuery(`name:"Hans" AND surname:"Wurst"`, &Node{
		NodeType: NodeAnd,
		Left: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Hans`},
			},
		},
		Right: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`surname`},
				Values: []string{`Wurst`},
			},
		},
	}))

	t.Run("simple AND clause with parenthesis", testParseQuery(`(name:"Hans" AND surname:"Wurst")`, &Node{
		NodeType: NodeAnd,
		Left: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Hans`},
			},
		},
		Right: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`surname`},
				Values: []string{`Wurst`},
			},
		},
	}))

	t.Run("simple OR clause", testParseQuery(`name:"Hans" OR name:"Siegfried"`, &Node{
		NodeType: NodeOr,
		Left: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Hans`},
			},
		},
		Right: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Siegfried`},
			},
		},
	}))

	t.Run("simple OR clause with parenthesis", testParseQuery(`(name:"Hans" OR name:"Siegfried")`, &Node{
		NodeType: NodeOr,
		Left: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Hans`},
			},
		},
		Right: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Siegfried`},
			},
		},
	}))

	t.Run("simple OR clause with parenthesis around condition", testParseQuery(`name:"Hans" OR (name:"Siegfried")`, &Node{
		NodeType: NodeOr,
		Left: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Hans`},
			},
		},
		Right: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Siegfried`},
			},
		},
	}))

	t.Run("simple AND clause with newline", testParseQuery("name:\"Hans\"\n\tAND surname:\"Wurst\"", &Node{
		NodeType: NodeAnd,
		Left: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Hans`},
			},
		},
		Right: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`surname`},
				Values: []string{`Wurst`},
			},
		},
	}))

	t.Run("simple OR clause with newline", testParseQuery("name:\"Hans\"\n\tOR surname:\"Wurst\"", &Node{
		NodeType: NodeOr,
		Left: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Hans`},
			},
		},
		Right: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`surname`},
				Values: []string{`Wurst`},
			},
		},
	}))

	t.Run("OR / AND clauses", testParseQuery(`name:"Hans" OR name:"Siegfried" AND age:9001`, &Node{
		NodeType: NodeOr,
		Left: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Hans`},
			},
		},
		Right: &Node{
			NodeType: NodeAnd,
			Left: &Node{
				NodeType: NodeTerminal,
				Comparison: Comparison{
					Op:     "==",
					Field:  []string{`name`},
					Values: []string{`Siegfried`},
				},
			},
			Right: &Node{
				NodeType: NodeTerminal,
				Comparison: Comparison{
					Op:     "==",
					Field:  []string{`age`},
					Values: []string{`9001`},
				},
			},
		},
	}))

	t.Run("OR / AND clauses reordered", testParseQuery(`name:"Hans" AND age:9001 OR name:"Siegfried"`, &Node{
		NodeType: NodeOr,
		Left: &Node{
			NodeType: NodeAnd,
			Left: &Node{
				NodeType: NodeTerminal,
				Comparison: Comparison{
					Op:     "==",
					Field:  []string{`name`},
					Values: []string{`Hans`},
				},
			},
			Right: &Node{
				NodeType: NodeTerminal,
				Comparison: Comparison{
					Op:     "==",
					Field:  []string{`age`},
					Values: []string{`9001`},
				},
			},
		},
		Right: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Siegfried`},
			},
		},
	}))

	t.Run("OR / AND clauses with paren precedence", testParseQuery(`(name:"Hans" AND age:9001) OR name:"Siegfried"`, &Node{
		NodeType: NodeOr,
		Left: &Node{
			NodeType: NodeAnd,
			Left: &Node{
				NodeType: NodeTerminal,
				Comparison: Comparison{
					Op:     "==",
					Field:  []string{`name`},
					Values: []string{`Hans`},
				},
			},
			Right: &Node{
				NodeType: NodeTerminal,
				Comparison: Comparison{
					Op:     "==",
					Field:  []string{`age`},
					Values: []string{`9001`},
				},
			},
		},
		Right: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`name`},
				Values: []string{`Siegfried`},
			},
		},
	}))

	t.Run("simple NOT clause", testParseQuery(`!name:"Hans" AND surname:"Wurst"`, &Node{
		NodeType: NodeAnd,
		Left: &Node{
			NodeType: NodeNot,
			Left: &Node{
				NodeType: NodeTerminal,
				Comparison: Comparison{
					Op:     "==",
					Field:  []string{`name`},
					Values: []string{`Hans`},
				},
			},
		},
		Right: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`surname`},
				Values: []string{`Wurst`},
			},
		},
	}))

	t.Run("alternate simple NOT clause", testParseQuery(`NOT name:"Hans" AND surname:"Wurst"`, &Node{
		NodeType: NodeAnd,
		Left: &Node{
			NodeType: NodeNot,
			Left: &Node{
				NodeType: NodeTerminal,
				Comparison: Comparison{
					Op:     "==",
					Field:  []string{`name`},
					Values: []string{`Hans`},
				},
			},
		},
		Right: &Node{
			NodeType: NodeTerminal,
			Comparison: Comparison{
				Op:     "==",
				Field:  []string{`surname`},
				Values: []string{`Wurst`},
			},
		},
	}))

	t.Run("simple NOT clause with parenthesis", testParseQuery(`!(name:"Hans" AND surname:"Wurst")`, &Node{
		NodeType: NodeNot,
		Left: &Node{
			NodeType: NodeAnd,
			Left: &Node{
				NodeType: NodeTerminal,
				Comparison: Comparison{
					Op:     "==",
					Field:  []string{`name`},
					Values: []string{`Hans`},
				},
			},
			Right: &Node{
				NodeType: NodeTerminal,
				Comparison: Comparison{
					Op:     "==",
					Field:  []string{`surname`},
					Values: []string{`Wurst`},
				},
			},
		},
	}))

	t.Run("float value", testParseQuery(`floppy:1.4`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`floppy`},
			Values: []string{`1.4`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("negative float value", testParseQuery(`floppy:-1.4`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`floppy`},
			Values: []string{`-1.4`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("int value", testParseQuery(`memory:32`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`memory`},
			Values: []string{`32`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("negative int value", testParseQuery(`memory:-32`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`memory`},
			Values: []string{`-32`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("boolean (true) value", testParseQuery(`isAdmin:true`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`isAdmin`},
			Values: []string{`true`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("boolean (false) value", testParseQuery(`writesGoodParsers:false`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`writesGoodParsers`},
			Values: []string{`false`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("cidr value", testParseQuery(`internal:192.168.1.0/24`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`internal`},
			Values: []string{`192.168.1.0/24`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("fullDate value", testParseQuery(`Andy:1979-10-03`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`Andy`},
			Values: []string{`1979-10-03`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("dateTime value", testParseQuery(`AndyPrecise:2021-06-08T20:56:33+00:00`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`AndyPrecise`},
			Values: []string{`2021-06-08T20:56:33+00:00`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("regexp value", testParseQuery(`domains:/.*\\.[a-z0-9]*\\.local/`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:    "==",
			Field: []string{`domains`},
			// TODO: Revisit when types are done to strip enclosing / (see aql.peg)
			// Values:  []string{`.*\\.[a-z0-9]*\\.local`},
			Values: []string{`/.*\\.[a-z0-9]*\\.local/`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("comparator == (implicit)", testParseQuery(`answer:42`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`answer`},
			Values: []string{`42`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("comparator ><", testParseQuery(`whiskers:><0`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "><",
			Field:  []string{`whiskers`},
			Values: []string{`0`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("comparator >", testParseQuery(`over9000:>9000`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     ">",
			Field:  []string{`over9000`},
			Values: []string{`9000`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("comparator >=", testParseQuery(`almost:>=9000`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     ">=",
			Field:  []string{`almost`},
			Values: []string{`9000`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("comparator <", testParseQuery(`alone:<2`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "<",
			Field:  []string{`alone`},
			Values: []string{`2`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("comparator <=", testParseQuery(`pair:<=2`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "<=",
			Field:  []string{`pair`},
			Values: []string{`2`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("allow leading whitespace", testParseQuery(` name:"Peter"`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`name`},
			Values: []string{`Peter`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("allow trailing whitespace", testParseQuery(`name:"Peter" `, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`name`},
			Values: []string{`Peter`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("allow whitespace before value", testParseQuery(`name: "Peter"`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`name`},
			Values: []string{`Peter`},
		},
		Left:  nil,
		Right: nil,
	}))

	t.Run("allow dash in field name", testParseQuery(`na-me: "Peter"`, &Node{
		NodeType: NodeTerminal,
		Comparison: Comparison{
			Op:     "==",
			Field:  []string{`na-me`},
			Values: []string{`Peter`},
		},
		Left:  nil,
		Right: nil,
	}))
}

func testParseQuery(query string, want *Node) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		t.Parallel()
		n, err := ParseQuery(query)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(want, n) {
			wb, _ := json.MarshalIndent(want, "", " ")
			nb, _ := json.MarshalIndent(n, "", " ")
			t.Fatalf("expected:\n%s\ngot:\n%s", string(wb), string(nb))
		}
	}
}
