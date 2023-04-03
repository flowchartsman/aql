package jsonmatcher

import "github.com/valyala/fastjson"

type exprNull struct {
	// stats here
}

// null will only return true if all values are explicitly null
func (m *exprNull) matches(field *field) bool {
	for v := range field.values {
		if field.values[v].Type() != fastjson.TypeNull {
			return false
		}
	}
	return true
}
