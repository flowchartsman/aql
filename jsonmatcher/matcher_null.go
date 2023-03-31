package jsonmatcher

import "github.com/valyala/fastjson"

type nullMatcher struct {
	// stats here
}

func (m *nullMatcher) matches(values []*fastjson.Value) bool {
	for v := range values {
		if values[v].Type() == fastjson.TypeNull {
			return true
		}
	}
	return false
}
