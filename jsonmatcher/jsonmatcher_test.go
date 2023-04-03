package jsonmatcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func ExampleMatcher() {
	const json = `{
		"date": "1970-01-02",
		"number": 2,
		"name": "Andy",
		"description": "大懒虫"
	}`

	m, err := NewMatcher(`name:"andy" AND date:><(1970-01-01,1980-01-01)`)
	if err != nil {
		log.Fatalf("error running query: %v", err)
	}

	result, err := m.Match([]byte(json))
	if err != nil {
		log.Fatalf("error during match: %v", err)
	}
	fmt.Println(result)

	// Output: true
}

func TestMatcher(t *testing.T) {
	jb, err := os.ReadFile(filepath.Join("testdata", "testdata.json"))
	if err != nil {
		t.Fatalf("failed to read json file: %v", err)
	}

	paths, err := filepath.Glob(filepath.Join("testData", "*.input"))
	if err != nil {
		t.Fatalf("test data not found")
	}
	for _, path := range paths {
		_, filename := filepath.Split(path)
		testname := filename[:len(filename)-len(filepath.Ext(path))]
		t.Run(testname, func(t *testing.T) {
			subtests, err := getTests(path)
			if err != nil {
				t.Fatalf("failed to read tests: %v", err)
			}
			for _, st := range subtests {
				t.Run(st.name, func(t *testing.T) {
					t.Logf(st.query)
					if st.skip {
						t.SkipNow()
					}
					if strings.TrimSpace(st.query) == "" {
						t.Fatalf("empty test")
					}
					matcher, err := NewMatcher(st.query)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					matched, err := matcher.Match(jb)
					if err != nil {
						t.Fatalf("unexpected matcher error: %v", err)
					}
					if matched != st.expect {
						t.Fatalf("want: %v, got: %v", st.expect, matched)
					}
				})
			}
		})
	}
}

type queryTest struct {
	expect bool
	name   string
	query  string
	skip   bool
	//file string
	//line string
}

var testMarker = regexp.MustCompile(`^[TFS] `)

func getTests(filename string) ([]queryTest, error) {
	var out []queryTest
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(file), "\n")
	var nextTest queryTest
	for lineno, line := range lines {
		if line == "" {
			continue
		}
		if testMarker.MatchString(line) {
			if nextTest.name != "" {
				return nil, fmt.Errorf("unexpected name at line %d", lineno+1)
			}
			name := line[2:]
			if name == "" {
				return nil, fmt.Errorf("empty name at line %d", lineno+1)
			}
			nextTest = queryTest{
				name: name,
			}
			switch line[0] {
			case 'T':
				nextTest.expect = true
			case 'F':
				nextTest.expect = false
			case 'S':
				nextTest.expect = false
				nextTest.skip = true
			}
			continue
		}

		if nextTest.name == "" {
			return nil, fmt.Errorf("unnamed test at line %d", lineno+1)
		}
		nextTest.query = line
		out = append(out, nextTest)
		nextTest = queryTest{}
	}
	return out, nil
}
