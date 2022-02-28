package token

import (
	"regexp"
	"strings"
)

type pattern struct {
	reStr     string
	tokenType Type
}

type classifier struct {
	re     *regexp.Regexp
	ctypes []Type
}

func newClassifier(patterns ...pattern) *classifier {
	ctypes := make([]Type, len(patterns))
	var reStr strings.Builder
	reStr.WriteString(`^(?:`)
	for i, p := range patterns {
		reStr.WriteString(`(`)
		reStr.WriteString(p.reStr)
		reStr.WriteString(`)`)
		if i < len(patterns)-1 {
			reStr.WriteString(`|`)
		}
		ctypes[i] = p.tokenType
	}
	reStr.WriteString(`)$`)
	return &classifier{
		re:     regexp.MustCompile(reStr.String()),
		ctypes: ctypes,
	}
}

func (c *classifier) getLiteralType(literal string) Type {
	m := c.re.FindStringSubmatch(literal)
	for i := 1; i < len(m); i++ {
		if m[i] != "" {
			return c.ctypes[i-1]
		}
	}
	return ILLEGAL
}
