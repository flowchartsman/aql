package jsonmatcher

import (
	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

type exprFuzzy struct {
	pat *search.Pattern
}

func newExprFuzzy(str string) *exprFuzzy {
	return &exprFuzzy{
		pat: search.New(language.Und, search.Loose).CompileString(str),
	}
}

func (e *exprFuzzy) matches(field *field) bool {
	for _, v := range field.scalarValues() {
		str, ok := getStringVal(v)
		if !ok {
			continue
		}
		start, _ := e.pat.IndexString(str)
		if start >= 0 {
			return true
		}
	}
	return false
}
