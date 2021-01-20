package jsonquery

import (
	"fmt"
	"strconv"

	"github.com/araddon/dateparse"
)

var opEQ = &equalityComparison{}

// this has gotta be re-done from the parser level up, these need types, this is dumb :)_
type equalityComparison struct{}

func (e *equalityComparison) GetComparator(rValues []string) (Comparator, error) {
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

var opNE = &negateComparison{opEQ}

type negateComparison struct {
	e *equalityComparison
}

func (ne *negateComparison) GetComparator(rValues []string) (Comparator, error) {
	ec, err := ne.e.GetComparator((rValues))
	if err != nil {
		return nil, err
	}
	return &invertComp{ec}, nil
}

type invertComp struct {
	c Comparator
}

func (i *invertComp) Evaluate(lValues []string) (bool, error) {
	b, err := i.c.Evaluate(lValues)
	if err == nil {
		b = !b
	}
	return b, err
}
