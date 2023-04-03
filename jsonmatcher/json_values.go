package jsonmatcher

import (
	"math"
	"strconv"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/valyala/fastjson"
)

type field struct {
	// offsets when possible
	values []*fastjson.Value
}

// TODO: cache paths
func getField(path []string, root *fastjson.Value) *field {
	values := getValues(path, root)
	return &field{
		values: values,
	}
}

func (f *field) scalarValues() []*fastjson.Value {
	var out []*fastjson.Value
	for _, v := range f.values {
		switch v.Type() {
		case fastjson.TypeObject:
			continue
		case fastjson.TypeArray:
			for _, av := range v.GetArray() {
				switch av.Type() {
				case fastjson.TypeObject, fastjson.TypeArray:
					continue
				default:
					out = append(out, av)
				}
			}
		default:
			out = append(out, v)
		}
	}
	return out
}

func (f *field) listValues() []*fastjson.Value {
	var out []*fastjson.Value
	for _, v := range f.values {
		if v.Type() == fastjson.TypeArray {
			out = append(out, v)
		}
	}
	return out
}

type valuepath []string

func (vp valuepath) bottom() bool {
	return len(vp) == 1
}

func (vp valuepath) next() valuepath {
	if vp.bottom() {
		return nil
	}
	return vp[1:]
}

func (vp valuepath) current() string {
	return vp[0]
}

// possible optimization: if a field is referenced only in an exists query,
// getValues can return early
func getValues(path valuepath, value *fastjson.Value) (fieldValues []*fastjson.Value) {
	var outputValues []*fastjson.Value

	switch value.Type() {
	case fastjson.TypeObject:
		// looking for a path segment in an object, so look for that key
		child := value.Get(path.current())
		if child == nil {
			return nil
		}
		// We found the value, and this is the last path segment, go ahead
		// and return it.
		if path.bottom() {
			return []*fastjson.Value{child}
		}
		// otherwise drill down on the next value in the chain
		outputValues = append(outputValues, getValues(path.next(), child)...)
	case fastjson.TypeArray:
		// looking for a path segment in an array, so look at every item, at
		// this same level.
		for _, child := range value.GetArray() {
			if child.Type() == fastjson.TypeObject {
				outputValues = append(outputValues, getValues(path, child)...)
			}
		}
	}
	return outputValues
}

// TODO: overload found to report type data for stat tracking
func getStringVal(v *fastjson.Value) (stringVal string, isStringy bool) {
	switch v.Type() {
	case fastjson.TypeString:
		return string(v.GetStringBytes()), true
	case fastjson.TypeNumber:
		return strconv.FormatFloat(v.GetFloat64(), 'f', -1, 64), true
	}
	return "", false
}

func getNumberVal(v *fastjson.Value) (floatVal float64, isNumeric bool) {
	switch v.Type() {
	case fastjson.TypeString:
		strVal := string(v.GetStringBytes())
		parsedFloat, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return math.NaN(), false
		}
		return parsedFloat, true
	case fastjson.TypeNumber:
		return v.GetFloat64(), true
	}
	return math.NaN(), false
}

func getBoolVal(v *fastjson.Value) (boolVal bool, found bool) {
	switch v.Type() {
	case fastjson.TypeFalse:
		return false, true
	case fastjson.TypeTrue:
		return true, true
	}
	return false, false
}

func getDatetimeVal(v *fastjson.Value) (intVal int64, found bool) {
	// TODO: tighten this up for numstrings. Probably want to be more careful
	// about what we consider a date with how flexible dateparse is. Maybe a
	// special getStringVal() that only accepts int-like unix epochs.
	sv, ok := getStringVal(v)
	if !ok {
		return 0, false
	}
	t, err := dateparse.ParseAny(sv)
	if err != nil {
		return 0, false
	}
	return t.UnixNano(), true
}

// true:
//   - <boolean> true
//   - <numeric> != 0
//   - <string> != ["", "false", "0"]
//
// false:
//   - boolean false
//   - <numeric> == 0
//   - <string> == ["", "false", "0"]
//   - null
func getTruthyVal(v *fastjson.Value) (boolVal bool, found bool) {
	switch v.Type() {
	case fastjson.TypeString:
		// all strings are truthy, except:
		//  - "" (empty string)
		//  - "0"
		//  - "false"
		sv := string(v.GetStringBytes())
		switch len(sv) {
		case 0:
			return false, true
		case 1:
			if sv == "0" {
				return false, true
			}
		case 5:
			if strings.ToLower(sv) == "false" {
				return false, true
			}
		}
		return true, true
	case fastjson.TypeNumber:
		// all numeric values are true, except for 0
		return v.GetFloat64() != 0, true
	case fastjson.TypeFalse:
		return false, true
	case fastjson.TypeTrue:
		return true, true
	case fastjson.TypeNull:
		// explicit null is false
		return false, true
	}
	return false, false
}
