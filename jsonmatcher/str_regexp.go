package jsonmatcher

import (
	"bytes"
	"regexp"
	"strings"
	"unicode/utf8"
)

// TODO: coalesce wcString ...string (combining string search regexp)
// TODO: detect crap wildcards like "foo**"
// esp w/unicode
// normalize spaces. fields->regexp
func stringSearchRegexp(wcString string) *regexp.Regexp {
	wcString = strings.ToLower(wcString)
	wcRunes := []rune(wcString)
	var buf bytes.Buffer
	isASCII := isASCII(wcString)
	buf.WriteString(`(?i)`)
	if isASCII {
		buf.WriteString(`\b`)
	}
	var accum bytes.Buffer
	for i := 0; i < len(wcRunes); i++ {
		switch wcRunes[i] {
		case '\\':
			if i < len(wcRunes)-1 {
				switch wcRunes[i+1] {
				case '?', '*':
					accum.WriteRune(wcRunes[i+1])
					i += 1
					continue
				}
			}
		case '?', '*':
			// escape everything until this point
			buf.WriteString(regexp.QuoteMeta(accum.String()))
			switch wcRunes[i] {
			case '?':
				buf.WriteString(`.`)
			case '*':
				buf.WriteString(`.*?`)
			}
			accum.Reset()
			continue
		}
		accum.WriteRune(wcRunes[i])
	}
	// write whatever remains in the accumulator
	buf.WriteString(regexp.QuoteMeta(accum.String()))
	if isASCII {
		buf.WriteString(`\b`)
	}
	return regexp.MustCompile(buf.String())
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
