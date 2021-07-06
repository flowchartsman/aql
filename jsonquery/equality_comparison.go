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
	var cs []Comparator
	if len(rValues) < 1 {
		return nil, fmt.Errorf("equality comparison needs more than one value")
	}
	// a little dirty to do it this way, but necessary to support `foo:("value", "value2")` in the current system
	// TODO: after types, comparators only take a single value struct w/type type-aware ops
	for _, rv := range rValues {
		if len(rv) == 0 {
			cs = append(cs, &stringEQComparator{""})
			continue
		}
		if _, err := strconv.ParseFloat(rv, 64); err == nil {
			c, err := opnumEQ.GetComparator([]string{rv})
			if err != nil {
				return nil, err
			}
			cs = append(cs, c)
			continue
		}
		if _, err := dateparse.ParseAny(rv); err == nil {
			c, err := opnumEQ.GetComparator([]string{rv})
			if err != nil {
				return nil, err
			}
			cs = append(cs, c)
			continue
		}
		if b, err := strconv.ParseBool(rv); err == nil {
			cs = append(cs, &boolEQComparator{b})
			continue
		}
		cs = append(cs, &stringEQComparator{rv})
	}
	return &multiComparator{cs}, nil
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

type multiComparator struct {
	cs []Comparator
}

func (m *multiComparator) Evaluate(lValues []string) (bool, error) {
	for _, c := range m.cs {
		match, err := c.Evaluate(lValues)
		if err != nil {
			return false, nil
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}
