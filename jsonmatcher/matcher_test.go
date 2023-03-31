package jsonmatcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
					matcher, _, err := NewMatcher(st.query)
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
	//file string
	//line string
}

func getTests(filename string) ([]queryTest, error) {
	var out []queryTest
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(file), "\n")
	name := ""
	expect := false
	for lineno, line := range lines {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "T ") || strings.HasPrefix(line, "F ") {
			if name != "" {
				return nil, fmt.Errorf("unexpected name at line %d", lineno+1)
			}
			expect = false
			if line[0] == 'T' {
				expect = true
			}
			name = line[2:]
		} else {
			if name == "" {
				return nil, fmt.Errorf("expected name at line %d", lineno+1)
			}
			out = append(out, queryTest{
				name:   name,
				expect: expect,
				query:  line,
			})
			name = ""
		}
	}
	return out, nil
}
