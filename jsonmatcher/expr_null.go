package jsonmatcher

import (
	"github.com/buger/jsonparser"
)

type exprNull struct {
	// stats here
}

// null will only return true if all values are explicitly null
func (m *exprNull) matches(field *field) bool {
	for i := range field.values {
		if field.values[i].dataType != jsonparser.Null {
			return false
		}
	}
	return true
}
