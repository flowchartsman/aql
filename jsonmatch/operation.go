package jsonmatch

import (
	"fmt"
)

// Operation represents a comparison operation
type Operation interface {
	// GetComparator returns a comparator for the operation
	GetComparator(rValues []string) (Comparator, error)
}

// Comparator is an instance of a particular operation with a particular set of
// rValues
type Comparator interface {
	// Evaluate checks all of the lvalues for validity and returns true if they
	// all match. If there is a problem with any of the lValues, it will return
	// false and an error
	Evaluate(lValues []string) (bool, error)
}

// Refreshable represents whether or not a comparator can be updated, like with
// a new relative time, for example
type Refreshable interface {
	// Refresh refreshes the comparator, returning an error if it cannot do so
	Refresh() error
}

// Arity is the number of rValues a comparison operation can be made with
type Arity int

// CheckArity verifies that the rvalues provided for the comparator match the operation arity
func (a Arity) CheckArity(rValues []string) error {
	if a < 0 {
		if len(rValues) == 0 {
			return fmt.Errorf("operator must have at least one value to compare against")
		}
	} else {
		if len(rValues) != int(a) {
			return fmt.Errorf("operator expects %d comparison values, but %d provided", a, len(rValues))
		}
	}
	return nil
}
