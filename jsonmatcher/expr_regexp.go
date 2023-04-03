package jsonmatcher

import (
	"regexp"
)

type exprRegexp struct {
	value *regexp.Regexp
}

func (e *exprRegexp) matches(field *field) bool {
	for _, v := range field.scalarValues() {
		str, ok := getStringVal(v)
		if !ok {
			continue
		}
		if e.value.MatchString(str) {
			return true
		}
	}
	return false
}
