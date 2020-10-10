package jsonmatch

import (
	"fmt"
	"time"
)

var opGT = &numericComparison{
	Arity: 1,
	comparisonFloat: func(lvals []float64, rval float64) bool {
		// will only ever be one
		return rval > lvals[0]
	},
	comparisonTime: func(lvals []time.Time, rval time.Time) bool {
		// will only ever be one
		return rval.After(lvals[0])
	},
}

var opGTE = &numericComparison{
	Arity: 1,
	comparisonFloat: func(lvals []float64, rval float64) bool {
		// will only ever be one
		return rval >= lvals[0]
	},
	comparisonTime: func(lvals []time.Time, rval time.Time) bool {
		// will only ever be one
		return rval.After(lvals[0]) || rval.Equal(lvals[0])
	},
}

var opLT = &numericComparison{
	Arity: 1,
	comparisonFloat: func(lvals []float64, rval float64) bool {
		// will only ever be one
		return rval < lvals[0]
	},
	comparisonTime: func(lvals []time.Time, rval time.Time) bool {
		// will only ever be one
		return rval.Before(lvals[0])
	},
}

var opLTE = &numericComparison{
	Arity: 1,
	comparisonFloat: func(lvals []float64, rval float64) bool {
		// will only ever be one
		return rval <= lvals[0]
	},
	comparisonTime: func(lvals []time.Time, rval time.Time) bool {
		// will only ever be one
		return rval.Before(lvals[0]) || rval.Equal(lvals[0])
	},
}

var opBetween = &numericComparison{
	Arity: 2,
	extraNumericValidator: func(rvals []float64) error {
		if rvals[0] >= rvals[1] {
			return fmt.Errorf("first value must be less than second value, but %f >= %f", rvals[0], rvals[1])
		}
		return nil
	},
	extraTimeValidator: func(rvals []time.Time) error {
		if rvals[0].After(rvals[1]) || rvals[0].Equal(rvals[1]) {
			return fmt.Errorf("first value must be before second value, but %s is after %s", rvals[0], rvals[1])
		}
		return nil
	},
	// TODO ops need optional validator
	comparisonFloat: func(lvals []float64, rval float64) bool {
		// will only ever be two
		return rval >= lvals[0] && rval <= lvals[1]
	},
	comparisonTime: func(lvals []time.Time, rval time.Time) bool {
		// will only ever be two
		return (rval.After(lvals[0]) || rval.Equal(lvals[0])) && (rval.Before(lvals[1]) || rval.Equal(lvals[1]))
	},
}

var opnumEQ = &numericComparison{
	Arity: 1,
	comparisonFloat: func(lvals []float64, rval float64) bool {
		for _, lv := range lvals {
			if rval == lv {
				return true
			}
		}
		return false
	},
	comparisonTime: func(lvals []time.Time, rval time.Time) bool {
		for _, lv := range lvals {
			if rval.Equal(lv) {
				return true
			}
		}
		return false
	},
}
