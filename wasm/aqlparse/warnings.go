package main

import (
	"fmt"
	"regexp/syntax"
	"strings"
	"unicode/utf8"

	"github.com/flowchartsman/aql/parser"
	"github.com/flowchartsman/aql/parser/ast"
)

func warningVisitor(node ast.Node, tape *parser.MessageTape) error {
	// TODO: this could deal with some untangling, it's getting quite long.
	// Maybe split up the checks more.
	switch n := node.(type) {
	case *ast.ExprNode:
		for _, v := range n.RVals {
			switch rv := v.(type) {
			case *ast.RegexpVal:
				if n.Op == ast.SIM {
					tape.WarningAt(n.Pos(), "Similarity comparison is no longer necessary for regular expressions and will eventually be removed. Please use the normal comparison operator - field:/<regular expression/")
				}
				reStr := strings.Trim(v.String(), `/`)
				if strings.HasPrefix(reStr, "(?i)") && !isASCII(v.String()) {
					tape.WarningAt(rv.Pos(), `Case-insensitive regular expression %s contains unicode characters. This may not work as intended. Consider a fuzzy match - 'field:~("value", "language")`, v.String())
				}
				reStr = strings.TrimPrefix(reStr, "(?i)")
				if strings.HasPrefix(reStr, ".*") || strings.HasSuffix(reStr, ".*") {
					tape.HintAt(rv.Pos(), "regular expression %s does not need to begin or end with \".*\", as this is redundant.", v.String())
				}
				reg, err := syntax.Parse(rv.Value().String(), syntax.Perl)
				if err != nil {
					return fmt.Errorf("unexpected error parsing regular expression: %v", err)
				}
				if len(reg.Sub) == 1 && reg.Op == syntax.OpCapture {
					tape.HintAt(rv.Pos(), `unnecessary outer capturing group "()", consider /%s/`, reg.Sub[0])
					reg = reg.Sub[0]
				}
				if reg.Op == syntax.OpAlternate && !doesSomethingSpecial(reg) {
					tape.HintAt(rv.Pos(), `if you are doing large string alternations, consider using a multi-string match: such as '%s:("one", "two")'`, strings.Join(n.Field, "."))
				}
				if len(reg.Sub) >= 2 {
					if isDotStar(reg.Sub[0]) {
						tape.HintAt(rv.Pos(), `leading ".*" may not do what you think, if you are searching for a term anywhere in a string, consider just "/word/"`)
					}
					if len(reg.Sub) >= 3 {
						if isDotStar(reg.Sub[len(reg.Sub)-1]) {
							tape.HintAt(rv.Pos(), `trailing ".*" is probably unnecessary`)
						}
					}
				}
			case *ast.TimeVal:
				if n.Op == ast.EQ {
					tape.WarningAt(rv.Pos(), `exact matches on datetime values currently match the time EXACTLY, consider using [:~] or a numeric comparison`)
				}
			}
		}
	}
	return nil
}

func doesSomethingSpecial(r *syntax.Regexp) bool {
	if len(r.Sub) == 1 {
		switch r.Op {
		case syntax.OpLiteral, syntax.OpStar, syntax.OpQuest:
			return false
		default:
			return true
		}
	} else {
		for i := range r.Sub {
			if doesSomethingSpecial(r.Sub[i]) {
				return true
			}
		}
	}
	return false
}

func isDotStar(r *syntax.Regexp) bool {
	if len(r.Sub) == 1 && r.Op == syntax.OpStar {
		if r.Sub[0].Op == syntax.OpAnyChar || r.Sub[0].Op == syntax.OpAnyCharNotNL {
			return true
		}
	}
	return false
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
