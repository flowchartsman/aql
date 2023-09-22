package jsonmatcher

import (
	"bytes"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

func getStringMatcher(str string) fieldExpr {
	if !isASCII(str) && !hasWildcard(str) {
		// TODO: replace this with a function type match
		// foo:lang_match(<lang>, <str>)
		// or
		// foo:lang_match(<str>) with auto-detect language
		return newUnicodeMatcher(str)
	}
	return &exprRegexp{
		value: stringSearchRegexp(str),
	}
}

// todo: replace runs of spaces or handle spaces specially somehow?
func stringSearchRegexp(wcString string) *regexp.Regexp {
	wcString = strings.ToLower(wcString)
	wcRunes := []rune(wcString)
	var buf bytes.Buffer
	buf.WriteString(`(?i)`)
	asciiOnly := isASCII(wcString)
	if asciiOnly {
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
	if asciiOnly {
		buf.WriteString(`\b`)
	}
	return regexp.MustCompile(buf.String())
}

type unicodeMatcher struct {
	pat *search.Pattern
}

func newUnicodeMatcher(str string) *unicodeMatcher {
	return &unicodeMatcher{
		pat: search.New(language.Und, search.Loose).CompileString(str),
	}
}

func (u *unicodeMatcher) matches(field *field) bool {
	for _, v := range field.scalarValues() {
		str, ok := getStringVal(v)
		if !ok {
			continue
		}
		start, _ := u.pat.IndexString(str)
		if start >= 0 {
			return true
		}
	}
	return false
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

func hasWildcard(s string) bool {
	return strings.IndexAny(s, "?*") >= 0
}
