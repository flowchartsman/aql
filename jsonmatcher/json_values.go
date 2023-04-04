package jsonmatcher

import (
	"math"
	"regexp"
	"strconv"

	"github.com/araddon/dateparse"
	"github.com/buger/jsonparser"
)

type jsonValue struct {
	data     []byte
	dataType jsonparser.ValueType
}

type field struct {
	// offsets when possible
	values []jsonValue
}

// TODO: cache paths
func getField(path []string, root []byte) *field {
	values := getValues(path, root, jsonparser.Object)
	return &field{
		values: values,
	}
}

func (f *field) scalarValues() []jsonValue {
	var out []jsonValue
	for _, v := range f.values {
		switch v.dataType {
		case jsonparser.Object:
			continue
		case jsonparser.Array:
			jsonparser.ArrayEach(v.data,
				func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					switch dataType {
					case jsonparser.Object, jsonparser.Array:
					default:
						out = append(out, jsonValue{
							data:     value,
							dataType: dataType,
						})
					}
				})
		default:
			out = append(out, v)
		}
	}
	return out
}

func (f *field) listValues() []jsonValue {
	var out []jsonValue
	for _, v := range f.values {
		if v.dataType == jsonparser.Array {
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
func getValues(path valuepath, data []byte, dataType jsonparser.ValueType) (fieldValues []jsonValue) {
	var outputValues []jsonValue

	switch dataType {
	case jsonparser.Object:
		// looking for a path segment in an object, so look for that key
		child, childType, _, _ := jsonparser.Get(data, path.current())
		if child == nil {
			return nil
		}
		// We found the value, and this is the last path segment, go ahead
		// and return it.
		if path.bottom() {
			return []jsonValue{{data: child, dataType: childType}}
		}
		// otherwise drill down on the next value in the chain
		outputValues = append(outputValues, getValues(path.next(), child, childType)...)
	case jsonparser.Array:
		// looking for a path segment in an array, so look at every item, at
		// this same level.
		jsonparser.ArrayEach(data,
			func(child []byte, dataType jsonparser.ValueType, offset int, err error) {
				if dataType == jsonparser.Object {
					outputValues = append(outputValues, getValues(path, child, jsonparser.Object)...)
				}
			})
	}
	return outputValues
}

// TODO: overload found to report type data for stat tracking
func getStringVal(v jsonValue) (stringVal string, isStringy bool) {
	switch v.dataType {
	case jsonparser.String:
		sv, err := jsonparser.ParseString(v.data)
		if err != nil {
			return "", false
		}
		return string(sv), true
	case jsonparser.Number:
		// TODO: unnecessary?
		fv, err := jsonparser.ParseFloat(v.data)
		if err != nil {
			return "", false
		}
		return strconv.FormatFloat(fv, 'f', -1, 64), true
	}
	return "", false
}

func getNumberVal(v jsonValue) (floatVal float64, isNumeric bool) {
	switch v.dataType {
	case jsonparser.String, jsonparser.Number:
		fv, err := jsonparser.ParseFloat(v.data)
		if err != nil {
			return math.NaN(), false
		}
		return fv, true
	}
	return math.NaN(), false
}

func getBoolVal(v jsonValue) (boolVal bool, found bool) {
	if v.dataType == jsonparser.Boolean {
		bv, err := jsonparser.ParseBoolean(v.data)
		if err != nil {
			return false, false
		}
		return bv, true
	}
	return false, false
}

func getDatetimeVal(v jsonValue) (intVal int64, found bool) {
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
var falsyString = regexp.MustCompile(`(?i)^(?:0|false)$`)

func getTruthyVal(v jsonValue) (boolVal bool, found bool) {
	switch v.dataType {
	case jsonparser.String:
		// all strings are truthy, except:
		//  - "" (empty string)
		//  - "0"
		//  - "false"
		if len(v.data) == 0 || falsyString.Match(v.data) {
			return false, true
		}
		return true, true
	case jsonparser.Number:
		fv, err := jsonparser.ParseFloat(v.data)
		if err != nil {
			return false, false
		}
		// all numeric values are true, except for 0
		return fv != 0, true
	case jsonparser.Boolean:
		bv, err := jsonparser.ParseBoolean(v.data)
		if err != nil {
			return false, false
		}
		return bv, true
	case jsonparser.Null:
		// explicit null is false
		return false, true
	}
	return false, false
}
