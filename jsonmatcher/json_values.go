package jsonmatcher

import (
	"math"
	"strconv"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/valyala/fastjson"
)

func getValues(path []string, parents ...*fastjson.Value) []*fastjson.Value {
	var testVals []*fastjson.Value
	for _, p := range parents {
		// get current path segment
		v := p.Get(path[0])
		if v == nil {
			// nothing found, return
			return nil
		}
		// found an array
		if v.Type() == fastjson.TypeArray {
			if len(path) == 1 {
				// if bottomed out, return all items in the array for comparisons
				testVals = append(testVals, v.GetArray()...)
			} else {
				// otherwise, recurse scan all items for the next path segment
				testVals = append(testVals, getValues(path[1:], v.GetArray()...)...)
			}
		} else {
			// found something else
			if len(path) == 1 {
				// if bottomed out, add it
				testVals = append(testVals, v)
			} else {
				// otherwise drill down
				testVals = append(testVals, getValues(path[1:], v)...)
			}
		}
	}
	return testVals
}

// TODO: overload found to report type data for stat tracking
func getStringVal(v *fastjson.Value) (stringVal string, isStringy bool) {
	switch v.Type() {
	case fastjson.TypeString:
		return string(v.GetStringBytes()), true
	case fastjson.TypeFalse:
		return "false", true
	case fastjson.TypeTrue:
		return "true", true
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
