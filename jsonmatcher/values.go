package jsonmatcher

import (
	"math"
	"strconv"

	"github.com/valyala/fastjson"
)

func getValues(path []string, parents ...*fastjson.Value) []*fastjson.Value {
	var testVals []*fastjson.Value
	for _, p := range parents {
		v := p.Get(path[0])
		if v == nil {
			return nil
		}
		if v.Type() == fastjson.TypeArray {
			if len(path) == 1 {
				testVals = append(testVals, v.GetArray()...)
			} else {
				testVals = append(testVals, getValues(path[1:], v.GetArray()...)...)
			}
		} else {
			if len(path) == 1 {
				testVals = append(testVals, v)
			} else {
				testVals = append(testVals, getValues(path[1:], v)...)
			}
		}
	}
	return testVals
}

func getValString(v *fastjson.Value) (stringVal string, isStringy bool) {
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

func getValNumeric(v *fastjson.Value) (floatVal float64, isNumeric bool) {
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
