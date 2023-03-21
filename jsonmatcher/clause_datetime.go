package jsonmatcher

import (
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	"github.com/flowchartsman/aql/parser/ast"
	"github.com/valyala/fastjson"
)

type datetimeClause struct {
	value time.Time
	op    ast.Op
}

func (d *datetimeClause) matches(values []*fastjson.Value) bool {
	for _, v := range values {
		sv, ok := getStringVal(v)
		if !ok {
			continue
		}
		t, err := dateparse.ParseAny(sv)
		if err != nil {
			// TODO: add in reporting with optional warn/fail for stuff like this
			// TODO: add invalid field reporting to field tracking stats
			return false
		}
		switch d.op {
		case ast.EQ:
			return t.Equal(d.value)
		case ast.LT:
			return t.Before(d.value)
		case ast.LTE:
			return t.Before(d.value) || t.Equal(d.value)
		case ast.GT:
			return t.After(d.value)
		case ast.GTE:
			return t.After(d.value) || t.Equal(d.value)
		case ast.SIM:
			// same day
			return t.Year() == d.value.Year() && t.YearDay() == d.value.YearDay()
		// need ast.BET for all valuetypes
		// case ast.BET:
		default:
			panic(fmt.Sprintf("invalid op for numeric comparison: %s", d.op))
		}
	}
	return false
}
