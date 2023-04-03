package jsonmatcher

type fieldExpr interface {
	matches(field *field) bool
}
