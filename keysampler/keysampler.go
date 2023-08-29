package keysampler

import (
	"bytes"
	"fmt"

	"github.com/buger/jsonparser"
)

type Stats struct {
	DocsScanned     int
	TotalKeys       int
	KeysAtThreshold int
}

// Sampler is a type that reads JSON documents and generates a list of
// keypaths derived from what it sees.
//
// It works by scanning JSON down to a certain depth, and tracking all keys that
// it traverses before bottoming out at a scalar value. If it encounters an
// array, it will traverse all items in the array transparently, however, it
// will skip arrays of arrays to prevent smashing the stack.
type Sampler struct {
	maxDepth  int
	keys      map[string]uint
	totalDocs int
}

func New(maxDepth int) *Sampler {
	k := &Sampler{
		maxDepth:  maxDepth,
		keys:      make(map[string]uint),
		totalDocs: 0,
	}
	return k
}

func (k *Sampler) Sample(sample []byte) error {
	var rootObj jsonparser.ValueType
	sampleLen := len(sample)
	if sampleLen > 100 {
		sampleLen = 100
	}
	switch sample[0] {
	case '{':
		if sample[len(sample)-1] != '}' {
			return fmt.Errorf("unterminated JSON object: %s", string(sample[:sampleLen]))
		}
		rootObj = jsonparser.Object
	case '[':
		if sample[len(sample)-1] != ']' {
			return fmt.Errorf("unterminated JSON array: %s", string(sample[:sampleLen]))
		}
		rootObj = jsonparser.Array
	default:
		return fmt.Errorf("sample is not an object or array: %s", string(sample[:sampleLen]))
	}
	// initial sampling state
	depth := 0
	path := &bytes.Buffer{} // todo: share
	err := k.sample(sample, path, depth, rootObj)
	if err != nil {
		k.keys = nil
		return err
	}
	k.totalDocs++
	return nil
}

func (k *Sampler) sample(current []byte, path *bytes.Buffer, depth int, valType jsonparser.ValueType) error {
	if depth > k.maxDepth {
		return nil
	}
	switch valType {
	case jsonparser.Object:
		pl := path.Len()
		return jsonparser.ObjectEach(current, func(key []byte, value []byte, dataType jsonparser.ValueType, _ int) error {
			path.WriteString(".")
			if bytes.ContainsAny(key, " ") {
				path.WriteString(`"`)
				path.Write(key)
				path.WriteString(`"`)
			} else {
				path.Write(key)
			}
			if err := k.sample(value, path, depth+1, dataType); err != nil {
				return err
			}
			path.Truncate(pl)
			return nil
		})
	case jsonparser.Array: // parse each array value
		// aerr represents any error we found deeper in the parse
		var aerr error
		_, err := jsonparser.ArrayEach(current, func(value []byte, dataType jsonparser.ValueType, _ int, ierr error) {
			// ierr represents the error from any previous array element; if
			// we've got one, return early
			if ierr != nil || aerr != nil {
				return
			}
			if dataType == jsonparser.Array {
				return
			}
			aerr = k.sample(value, path, depth, dataType)
		})
		if err == nil {
			err = aerr
		}
		return err
	case jsonparser.Unknown:
		return fmt.Errorf("unkown JSON value type: %v", string(current))
	default:
		// TODO Track types found for UI hinting
		k.keys[path.String()[1:]]++
	}
	return nil
}

func (k *Sampler) Keys() []string {
	out := make([]string, 0, len(k.keys))
	for key := range k.keys {
		out = append(out, key)
	}
	return out
}

func (k *Sampler) KeysAtThreshold(thresholdPercent int) []string {
	out := []string{}
	pct := float64(thresholdPercent) / 100
	for key, count := range k.keys {
		if float64(count)/float64(k.totalDocs) >= pct {
			out = append(out, key)
		}
	}
	return out
}

func (k *Sampler) Scanned() int {
	return k.totalDocs
}
