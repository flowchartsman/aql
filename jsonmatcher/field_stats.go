package jsonmatcher

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/valyala/fastjson"
	"go.uber.org/atomic"
)

const (
	fieldMarkWindow  = 1000
	fieldNumExamples = 10
	fieldMapMaxKeys  = 10
)

type expectedType string

const (
	expectString  = "string"
	expectNumeric = "numeric"
	expectBoolean = "boolean"
	expectNull    = "null"
	expectExists  = "<exists>"
)

const (
	str = iota
	strInArray
	number
	numberInArray
	object
	objectInArray
	boolean
	booleanInArray
	array
	arrayInArray
	null
	nullInArray
	numEncounteredTypes
)

var encounteredName = [numEncounteredTypes]string{
	"string",
	"string (in array)",
	"number",
	"number (in array)",
	"object",
	"object (in array)",
	"boolean",
	"boolean (in array)",
	"array",
	"array (in array)",
	"null",
	"null (in array)",
}

// TODO: Replace uber types with native types pending MarshalScalar (see field_stats.go)
type FieldStats struct {
	Expecting       []expectedType                        `json:"expecting"`
	TimesSampled    atomic.Int64                          `json:"times_sampled"`
	TypesEnountered [numEncounteredTypes]EncounteredStats `json:"types_encountered"`
}

func NewFieldStats(expectedTypes ...expectedType) *FieldStats {
	fs := &FieldStats{
		Expecting: expectedTypes,
	}
	for _, enc := range fs.TypesEnountered {
		enc.Examples = &ExampleList{}
	}
	return &FieldStats{
		Expecting: expectedTypes,
	}
}

func (n *FieldStats) MarshalJson() ([]byte, error) {
	out := map[string]interface{}{}
	out["expecting"] = n.Expecting
	out["times_seen"] = n.TimesSampled.Load()
	encountered := map[string]map[string]interface{}{}
	for i, e := range n.TypesEnountered {
		if e.TimesSeen.Load() != 0 {
			em := map[string]interface{}{}
			em["times_seen"] = e.TimesSeen.Load()
			em["times_matched"] = e.TimesMatched.Load()
			em["examples"] = e.Examples.Get()
			encountered[encounteredName[i]] = em
		}
	}
	if len(encountered) > 0 {
		out["encountered"] = encountered
	}
	return json.Marshal(out)
}

func (n *FieldStats) mark(foundValues []*fastjson.Value, matchIdx int, inArray bool) {
	n.TimesSampled.Inc()
	for _, foundValue := range foundValues {
		var which int
		var isObject bool
		switch foundValue.Type() {
		case fastjson.TypeString:
			which = str
		case fastjson.TypeNumber:
			which = number
		case fastjson.TypeFalse, fastjson.TypeTrue:
			which = boolean
		case fastjson.TypeArray:
			which = array
		case fastjson.TypeObject:
			which = object
			isObject = true
		case fastjson.TypeNull:
			which = null
		}
		if inArray {
			which++
		}
		n.TypesEnountered[which].TimesSeen.Inc()
		// every 1000 times this field type marked, extract an example field
		if which < boolean && n.TimesSampled.Load()%fieldMarkWindow == 0 {
			if !isObject {
				n.TypesEnountered[which].Examples.addExample(foundValue.String())
			} else {
				var sb strings.Builder
				sb.WriteString("<object with keys: ")
				numkeys := 0
				keylist := []string{}
				obj := foundValue.GetObject()
				obj.Visit(func(key []byte, _ *fastjson.Value) {
					if numkeys > fieldMapMaxKeys {
						return
					}
					keylist = append(keylist, `"`+string(key)+`"`)
				})
				sb.WriteString(strings.Join(keylist, `,`))
				if numkeys > fieldMapMaxKeys {
					sb.WriteString(", ...")
				}
				sb.WriteString(">")
				n.TypesEnountered[which].Examples.addExample(sb.String())
			}
		}
	}
}

type EncounteredStats struct {
	TimesSeen    atomic.Int64 `json:"times_seen"`
	TimesMatched atomic.Int64 `json:"times_matched"`
	Examples     *ExampleList `json:"examples,omitempty"`
}

type ExampleList struct {
	mux      sync.Mutex
	examples [fieldNumExamples]string
	eIdx     int
}

func (el *ExampleList) addExample(example string) {
	el.mux.Lock()
	defer el.mux.Unlock()
	el.eIdx++
	if el.eIdx == len(el.examples) {
		el.eIdx = 0
	}
	el.examples[el.eIdx] = example
}

func (el *ExampleList) Get() []string {
	el.mux.Lock()
	defer el.mux.Unlock()
	out := []string{}
	for _, example := range el.examples {
		if example == "" {
			break
		}
		out = append(out, example)
	}
	return out
}
