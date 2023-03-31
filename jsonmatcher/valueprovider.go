package jsonmatcher

import "time"

type ValueType int

const (
	unknown ValueType = iota
	Bool
	Float
	Int
	String
	Date
	Null
)

type Val interface {
	Type() ValueType
	Bool() (bool, bool)
	String() (string, bool)
	Float() (float64, bool)
	Int() (int, bool)
	Date() (time.Time, bool)
}

// TODO: research implementing Impl [JSON]ValueProvider
type VelueProvider interface {
	Get(path []string) (values []Val, found bool)
}
