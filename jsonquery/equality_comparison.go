package jsonquery

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/araddon/dateparse"
)

var opEQ = &equalityComparison{}

// this has gotta be re-done from the parser level up, these need types, this is dumb :)_
type equalityComparison struct{}

func (e *equalityComparison) GetComparator(rValues []string) (Comparator, error) {
	// This function attempts to figure out what kind of string has come out of the parser, even though
	// the parser can already tell us a lot of type information. Pending integration of these types in the next version
	// that means this function handles most of the string massaging into types. An unfortunate side effect of this is that
	// currently there's no way to distinguish between
	// foo:true
	// and
	// foo:"true"
	// When the former should be a boolean match, and the latter should be a string.
	if len(rValues) != 1 {
		return nil, fmt.Errorf("too many values for equality comparison. want 1, got %d", len(rValues))
	}
	if len(rValues[0]) == 0 {
		return nil, fmt.Errorf("cannot use empty identifier for equality comparison")
	}
	// this here is why we need type annotations in the parser
	if _, err := strconv.ParseFloat(rValues[0], 64); err == nil {
		return opnumEQ.GetComparator(rValues)
	}
	if _, err := dateparse.ParseAny(rValues[0]); err == nil {
		return opnumEQ.GetComparator(rValues)
	}
	if b, err := strconv.ParseBool(rValues[0]); err == nil {
		return &boolEQComparator{b}, nil
	}
	return &stringEQComparator{rValues[0]}, nil
}

type boolEQComparator struct {
	rval bool
}

func (b *boolEQComparator) Evaluate(lValues []string) (bool, error) {
	for _, bs := range lValues {
		bv, err := strconv.ParseBool(bs)
		if err != nil {
			return false, fmt.Errorf("non-boolean value found: %s", bs)
		}
		if bv == b.rval {
			return true, nil
		}
	}
	return false, nil
}

type stringEQComparator struct {
	rval string
}

func (s *stringEQComparator) Evaluate(lValues []string) (bool, error) {
	for _, rs := range lValues {
		if rs == s.rval {
			return true, nil
		}
	}
	return false, nil
}

type regexpEQComparator struct {
	rex *regexp.Regexp
}

func (r *regexpEQComparator) Evaluate(lValues []string) (bool, error) {
	for _, rs := range lValues {
		if r.rex.MatchString(rs) {
			return true, nil
		}
	}
	return false, nil
}
