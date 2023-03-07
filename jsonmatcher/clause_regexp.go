package jsonmatcher

import (
	"regexp"

	"github.com/valyala/fastjson"
)

type regexpClause struct {
	value *regexp.Regexp
}

func (r *regexpClause) matches(values []*fastjson.Value) bool {
	for _, v := range values {
		str, ok := getValString(v)
		if !ok {
			continue
		}
		if r.value.MatchString(str) {
			return true
		}
	}
	return false
}
