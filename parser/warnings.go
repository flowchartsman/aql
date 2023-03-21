package parser

import (
	"fmt"
	"strings"

	"github.com/flowchartsman/aql/parser/ast"
	"golang.org/x/exp/utf8string"
)

type warner struct {
	withWarnings bool
	warnings     []*ParseWarning
}

func newWarner(withWarnings bool) *warner {
	return &warner{
		withWarnings: withWarnings,
	}
}

func (w *warner) addWarning(node ast.Node, warnmsg string, v ...any) {
	// TODO: add location to node so it can be attached to warning-wrapped ParseError.
	// ParseWarning(newParseError(node))
	w.warnings = append(w.warnings, &ParseWarning{
		Inner:  fmt.Errorf(warnmsg, v...),
		Line:   -1,
		Column: -1,
		Offset: -1,
	})
}

func (w *warner) Visit(node ast.Node) ast.Visitor {
	if !w.withWarnings && len(w.warnings) > 0 {
		return nil
	}
	switch n := node.(type) {
	case *ast.ExprNode:
		for _, v := range n.RVals {
			switch val := v.(type) {
			case *ast.RegexpVal:
				if n.Op == ast.SIM {
					w.addWarning(node, "similarity comparison is no longer necessary for regular expressions and will eventually be removed. Please use the normal comparison operator (:).")
				}
				reStr := strings.Trim(val.ValStr(), "/")
				// TODO: identify good candidates for string match instead of
				// /(?i)foo|bar|baz/
				// may need to inspect val.Value() regexp value
				reStr = strings.TrimPrefix(reStr, "(?i)")
				if strings.HasPrefix(reStr, ".*") || strings.HasSuffix(reStr, ".*") {
					w.addWarning(node, "regular expression %s does not need to begin or end with \".*\", as this is redundant. Use /%s/.", val.ValStr(), strings.TrimPrefix(strings.TrimSuffix(reStr, ".*"), ".*"))
				}
				if !utf8string.NewString(val.ValStr()).IsASCII() {
					w.addWarning(node, "regular expression %s contains unicode characters. This will not work as intended for case-insensitive matching. Consider a string match.", val.ValStr())
				}
			}
		}
		return nil
	}
	return w
}
