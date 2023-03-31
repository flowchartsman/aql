package jsonmatcher

import "github.com/valyala/fastjson"

type existsMatcher struct {
	// stats here
}

func (m *existsMatcher) matches(values []*fastjson.Value) bool {
	return len(values) > 0
}
