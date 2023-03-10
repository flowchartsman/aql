package jsonmatcher

import (
	"bytes"
	"regexp"
	"strings"
)

func wildcardRegexp(wcString string) *regexp.Regexp {
	wcString = strings.ToLower(wcString)
	wcRunes := []rune(wcString)
	var buf bytes.Buffer
	buf.WriteString(`(?i)`)
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
			buf.WriteString(regexp.QuoteMeta(accum.String()))
			switch wcRunes[i] {
			case '?':
				buf.WriteString(`.`)
			case '*':
				buf.WriteString(`.*`)
			}
			accum.Reset()
			continue
		}
		accum.WriteRune(wcRunes[i])
	}
	buf.WriteString(regexp.QuoteMeta(accum.String()))
	return regexp.MustCompile(buf.String())
}

func wordRegexp(str string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b` + strings.ToLower(str) + `\b`)
}
