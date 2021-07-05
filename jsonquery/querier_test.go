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
	},
	"attributes": {
		"nice": true,
		"dead": false,
		"fun":"true",
		"enjoys_excessive_tests":"false",
		"likes_binary": 1,
		"killer_robot": 0,
		"ternary": 3
	},
	"net": {
		"router": "192.168.1.0"
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

func TestSimilarityMatch(t *testing.T) {
	t.Run("regexp prefix match", testJSQuery(`text.name:~/^And/`, true))
	t.Run("regexp suffix match", testJSQuery(`text.name:~/dy$/`, true))
	t.Run("regexp basic match", testJSQuery(`text.name:~/nd/`, true))
	t.Run("regexp case-insensitive match", testJSQuery(`text.name:~/(?i)^andy/`, true))
	t.Run("regexp case-sensitive match fail", testJSQuery(`text.name:~/^andy/`, false))
	t.Run("regexp unicode alias", testJSQuery(`text.description:~/\p{Han}{3}/`, true))
	t.Run("regexp escaped frontslashes", testJSQuery(`date.shortDate:~/^\d+\/\d+\/\d+/`, true))
	t.Run("wildcard prefix", testJSQuery(`text.likes:~"Pizza*"`, true))
	t.Run("wildcard suffix", testJSQuery(`text.likes:~"*(dog)"`, true))
	t.Run("truthy true bool", testJSQuery(`attributes.nice:~true`, true))
	t.Run("truthy true string", testJSQuery(`attributes.fun:~true`, true))
	t.Run("truthy true int", testJSQuery(`attributes.likes_binary:~true`, true))
	t.Run("truthy false bool", testJSQuery(`attributes.dead:~false`, true))
	t.Run("truthy false string", testJSQuery(`attributes.enjoys_excessive_tests:~false`, true))
	t.Run("truthy false int", testJSQuery(`attributes.killer_robot:~false`, true))
	t.Run("non-truthy isn't true", testJSQuery(`attributes.ternary:~true`, false))
	t.Run("non-truthy isn't false", testJSQuery(`attributes.ternary:~false`, false))
	t.Run("valid net contains", testJSQuery(`net.router:~192.168.1.0/24`, true))
	t.Run("valid net does not contain", testJSQuery(`net.router:~192.168.2.0/24`, false))
	t.Run("invalid address not contained", testJSQuery(`text.name:~192.168.2.0/24`, false))
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
