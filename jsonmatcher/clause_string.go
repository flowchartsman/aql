package jsonmatcher

import (
	"github.com/valyala/fastjson"
)

type stringClause struct {
	value string
}

func (s *stringClause) matches(values []*fastjson.Value) bool {
	for _, v := range values {
		str, ok := getStringVal(v)
		if !ok {
			continue
		}
		if str == s.value {
			return true
		}
	}
	return false
}
