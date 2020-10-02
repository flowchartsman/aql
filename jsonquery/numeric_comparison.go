package jsonquery

import (
	"fmt"
	"strconv"
	"time"

	"github.com/araddon/dateparse"
)

type numericComparison struct {
	Arity
	// TODO: SWAP ORDER
	comparisonFloat       func(rValues []float64, lValue float64) bool
	comparisonTime        func(rValues []time.Time, lvalue time.Time) bool
	extraNumericValidator func(rValues []float64) error
	extraTimeValidator    func(rValues []time.Time) error
}

func (n *numericComparison) GetComparator(rValues []string) (Comparator, error) {
	if err := n.CheckArity(rValues); err != nil {
		return nil, err
	}
	// attempt to detect if we're being called in a time context by testing the first lvalue
	if _, err := dateparse.ParseAny(rValues[0]); err != nil {
		// doesn't look like a time, assume float context
		floatValues, err := getFloatSlice(rValues)
		if err != nil {
			return nil, fmt.Errorf("numeric comparison value: %w", err)
		}
		if n.extraNumericValidator != nil {
			if err := n.extraNumericValidator(floatValues); err != nil {
				return nil, fmt.Errorf("invalid comparison values: %w", err)
			}
		}
		return &floatComparator{
			rValues:  floatValues,
			compFunc: n.comparisonFloat,
		}, nil
	}
	timeValues, err := getTimeSlice(rValues)
	if err != nil {
		return nil, fmt.Errorf("datetime comparison: %w", err)
	}
	if n.extraTimeValidator != nil {
		if err := n.extraTimeValidator(timeValues); err != nil {
			return nil, fmt.Errorf("invalid comparison values: %w", err)
		}
	}
	return &timeComparator{
		rValues:  timeValues,
		compFunc: n.comparisonTime,
	}, nil
}

type timeComparator struct {
	rValues  []time.Time
	compFunc func(l []time.Time, r time.Time) bool
}

func (t *timeComparator) Evaluate(lValues []string) (bool, error) {
	lTimevals, err := getTimeSlice(lValues)
	if err != nil {
		return false, fmt.Errorf("non-datetime value cannot be compared: %w", err)
	}
	for _, tv := range lTimevals {
		if t.compFunc(t.rValues, tv) {
			return true, nil
		}
	}
	return false, nil
}

type floatComparator struct {
	rValues  []float64
	compFunc func(r []float64, l float64) bool
}

func (f *floatComparator) Evaluate(lValues []string) (bool, error) {
	lFloatvals, err := getFloatSlice(lValues)
	if err != nil {
		return false, fmt.Errorf("non-numeric value cannot be compared: %w", err)
	}
	for _, lv := range lFloatvals {
		if f.compFunc(f.rValues, lv) {
			return true, nil
		}
	}
	return false, nil
}

func getFloatSlice(values []string) ([]float64, error) {
	floatValues := make([]float64, len(values))
	for i, fs := range values {
		fv, err := strconv.ParseFloat(fs, 64)
		if err != nil {
			return nil, fmt.Errorf("%q is not numeric", fs)
		}
		floatValues[i] = fv
	}
	return floatValues, nil
}

func getTimeSlice(values []string) ([]time.Time, error) {
	timeValues := make([]time.Time, len(values))
	for i, ts := range values {
		tv, err := dateparse.ParseAny(ts)
		if err != nil {
			return nil, fmt.Errorf("%q is not a recognized time format", ts)
		}
		timeValues[i] = tv
	}
	return timeValues, nil
}
