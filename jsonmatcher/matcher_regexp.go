package jsonmatcher

import (
	"regexp"

	"github.com/valyala/fastjson"
)

type regexpMatcher struct {
	value *regexp.Regexp
}

func (r *regexpMatcher) matches(values []*fastjson.Value) bool {
	for _, v := range values {
		str, ok := getStringVal(v)
		if !ok {
			continue
		}
		if r.value.MatchString(str) {
			return true
		}
	}
	return false
}
