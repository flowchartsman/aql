package jsonquery

import (
	"bytes"
	"testing"
)

const jsondoc = `{
	"date" : {
		"fullDate" : "1970-01-02",
		"dateTime" : "1970-01-02T15:53:33+00:00",
		"shortDate": "1/2/70",
		"writtenDate" : "January 1st, 1970"
	},
	"text" : {
		"name" : "Andy",
		"description": "大懒虫",
		"likes" : [
			"Pizza (food)",
			"Pig (dog)",
			"辣椒油 (food)",
			"Rolf (dog)",
			"Kirby (cat)"
		]
	},
	"number" : {
		"int": 1,
		"float": 1.1,
		"not": "hello"
	}
}`

func TestExistsMatch(t *testing.T) {
	t.Run("exists", testJSQuery(`text.name:exists`, true))
	t.Run("not exists", testJSQuery(`!text.greatest_fear:exists`, true))
}

func TestDateMatch(t *testing.T) {
	t.Run("fullDate < fullDate", testJSQuery(`date.fullDate:<1980-01-01`, true))
	t.Run("fullDate > fullDate", testJSQuery(`date.fullDate:>1970-01-01`, true))
	t.Run("dateTime < fullDate", testJSQuery(`date.dateTime:<1980-01-01`, true))
	t.Run("shortDate < fullDate", testJSQuery(`date.shortDate:<1980-01-01`, true))
	t.Run("dateTime < dateTime", testJSQuery(`date.dateTime:<1970-01-02T15:53:34+00:00`, true))
	t.Run("writtenDate < fullDate", testJSQuery(`date.writtenDate:<1970-02-01`, true))
	t.Run("!fullDate > fullDate", testJSQuery(`date.fullDate:>1980-01-01`, false))
	t.Run("date between", testJSQuery(`date.fullDate:><[1970-01-01,1970-01-03]`, true))
	t.Run("!date between", testJSQuery(`date.fullDate:><[1970-01-03,1970-01-04]`, false))
}

func TestRegexMatch(t *testing.T) {
	t.Run("prefix match", testJSQuery(`text.name:/^And/`, true))
	t.Run("suffix match", testJSQuery(`text.name:/dy$/`, true))
	t.Run("basic match", testJSQuery(`text.name:/nd/`, true))
	t.Run("case-insensitive match", testJSQuery(`text.name:/(?i)^andy/`, true))
	t.Run("!case-insensitive match", testJSQuery(`text.name:/^andy/`, false))
	t.Run("unicode alias", testJSQuery(`text.description:/\p{Han}{3}/`, true))
	t.Run("escaped frontslashes", testJSQuery(`date.shortDate:/^\d+\/\d+\/\d+/`, true))

	// t.Run("unicode alias", testJSQuery(`text.likes:/\p{Han}.*\(food\)$/`, true))
}

func TestNumericMatch(t *testing.T) {
	t.Run("eq", testJSQuery(`number.int:1`, true))
	t.Run("gt", testJSQuery(`number.int:>0`, true))
	t.Run("!gt", testJSQuery(`number.int:>1`, false))
	t.Run("ge", testJSQuery(`number.int:>=0`, true))
}

func testJSQuery(query string, want bool) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		t.Parallel()
		q, err := NewQuerier(query)
		if err != nil {
			t.Fatalf("unexpected query parse error: %v", err)
		}
		r := bytes.NewReader([]byte(jsondoc))
		result, err := q.Match(r)
		if err != nil {
			t.Fatalf("unexpected match error: %v", err)
		}
		switch want {
		case true:
			if !result {
				t.Fatalf("failed to match")
			}
		case false:
			if result {
				t.Fatalf("matched incorrectly")
			}
		}
	}
}
