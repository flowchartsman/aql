package jsonmatcher

type exprExists struct {
	// stats here
}

func (e *exprExists) matches(field *field) bool {
	return len(field.values) > 0
}
