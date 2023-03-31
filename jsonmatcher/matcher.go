package jsonmatcher

import (
	"github.com/valyala/fastjson"
)

type matcher interface {
	matches(values []*fastjson.Value) bool
}
