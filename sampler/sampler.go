package sampler

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

// New creates a new sampler that will extract key paths down to maxDepth depth
// (<0) means no limit.
func New(maxDepth int) *Sampler {
	k := &Sampler{
		maxDepth:  maxDepth,
		keys:      make(map[string]uint),
		totalDocs: 0,
	}
	return k
}

func (s *Sampler) Sample(sample []byte) error {
	var rootObj jsonparser.ValueType
	sampleLen := len(sample)
	// prep to get 100 characters or so to print if it looks weird, so that
	// there's something to examine in the logs
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
	// prep initial sampling state
	depth := 0
	path := &bytes.Buffer{} // todo: share
	err := s.sample(sample, path, depth, rootObj)
	if err != nil {
		s.keys = nil
		return err
	}

	s.totalDocs++
	return nil
}

func (s *Sampler) sample(current []byte, path *bytes.Buffer, depth int, valType jsonparser.ValueType) error {
	if depth > s.maxDepth {
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
			if err := s.sample(value, path, depth+1, dataType); err != nil {
				return err
			}
			path.Truncate(pl)
			return nil
		})
	case jsonparser.Array: // parse each array value
		// aerr represents any error we found deeper in the array parse
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
			aerr = s.sample(value, path, depth, dataType)
		})
		if err == nil {
			err = aerr
		}
		return err
	case jsonparser.Unknown:
		return fmt.Errorf("unkown JSON value type: %v", string(current))
	default:
		// TODO Track types found for UI hinting
		s.keys[path.String()[1:]]++
	}
	return nil
}

func (s *Sampler) Keys() []string {
	out := make([]string, 0, len(s.keys))
	for key := range s.keys {
		out = append(out, key)
	}
	return out
}

func (s *Sampler) KeysAtThreshold(thresholdPercent int) []string {
	out := []string{}
	pct := float64(thresholdPercent) / 100
	for key, count := range s.keys {
		if float64(count)/float64(s.totalDocs) >= pct {
			out = append(out, key)
		}
	}
	return out
}

func (s *Sampler) Reset() {
	// clear(k.keys)
	for k := range s.keys {
		delete(s.keys, k)
	}
	s.totalDocs = 0
}

func (s *Sampler) Scanned() int {
	return s.totalDocs
}
