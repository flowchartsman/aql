package jsonmatcher

import (
	"github.com/valyala/fastjson"
	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

type fuzzyClause struct {
	pat *search.Pattern
}

func newFuzzyClause(str string) *fuzzyClause {
	return &fuzzyClause{
		pat: search.New(language.Und, search.Loose).CompileString(str),
	}
}

func (s *fuzzyClause) matches(values []*fastjson.Value) bool {
	for _, v := range values {
		str, ok := getStringVal(v)
		if !ok {
			continue
		}
		start, _ := s.pat.IndexString(str)
		if start >= 0 {
			return true
		}
	}
	return false
}
