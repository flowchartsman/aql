package jsonmatcher

import (
	"github.com/valyala/fastjson"
)

type clause interface {
	matches(values []*fastjson.Value) bool
}
