package sampler

import (
	"sort"
	"strings"
	"testing"
)

func TestKeySampler(t *testing.T) {
	const (
		doc1 = `{
			"a": {
				"a1": "a1",
				"a2": {
					"a21": 1,
					"a22": ["a221"],
					"a23": {
						"a231": "a231"
					},
					"a24":[[]]
				}
			}
		}`
		doc2 = `{
			"a":{
				"a2": {
					"a24":1
				},
				"a3":"a3",
				"a4": {
					"a41": {
						"a411": {
							"a4111":"a4111",
							"a4112": {
								"a41121": {
									"a411211": "a411211"
								}
							}
						}
					}
				}
			}
		}`
		doc3 = `{
			"b": {
				"b1": {
					"b11": 1
				},
				"b2": "b2",
				"b3": {
					"b31": "b311"
				}
			}
		}`
	)
	k := New(5)

	for i := 0; i < 100; i++ {
		var err error
		if i%2 == 0 {
			if i%10 == 0 {
				err = k.Sample([]byte(doc3))
			} else {
				err = k.Sample([]byte(doc1))
			}
		} else {
			err = k.Sample([]byte(doc2))
		}
		if err != nil {
			t.Logf("unexpected sampling error - %v", err)
			t.Fail()
			break
		}
	}
	aKeys := k.Keys()
	sort.Strings(aKeys)
	tKeys := k.KeysAtThreshold(33)
	sort.Strings(tKeys)
	if k.Scanned() != 100 {
		t.Logf("Scanned() - expected: 100 got: %d", k.Scanned())
		t.Fail()
	}
	if len(k.Keys()) != 10 {
		t.Logf("len(Keys()) - expected: 10 got: %d", len(k.keys))
		t.Logf("keys: %q", aKeys)
		t.Fail()
	}
	if len(tKeys) != 7 {
		t.Logf("KeysAtThreshold(33) - expected: 7 got: %d", len(tKeys))
		t.Logf("keys: %q", tKeys)
		t.Fail()
	}
}

type docTest struct {
	doc          string
	depth        int
	expectedKeys []string
	expectedErr  string
}

func TestKeyExtraction(t *testing.T) {
	testDocCase(t, docTest{
		doc: `{
			"a":{
				"a1": {
					"a11":1
				},
				"a2":"a2v",
				"a3": {
					"a31": {
						"a311": {
							"a 3111":"a3111v",
							"a3112": {
								"a31121": {
									"a311211": 0
								}
							}
						}
					}
				}
			}
		}`,
		depth:        5,
		expectedKeys: []string{"a.a1.a11", "a.a2", `a.a3.a31.a311."a 3111"`},
	})
	testDocCase(t, docTest{
		doc: `{
				"a": [
					{
						"a1":"a1v"
					},
					{
						"a2":"a2v"
					},
					{
						"a3":"a3v"
					},
					{
						"a4":"a4v",
						"a5":"a5v"
					}
				]	
		}`,
		depth: 5,
		expectedKeys: []string{
			"a.a1", "a.a2", "a.a3", "a.a4", "a.a5",
		},
	})
	testDocCase(t, docTest{
		doc: `[
				{
					"b1":"b1v"
				},
				{
					"b2":"b2v"
				},
				{
					"b3":"b3v"
				},
				{
					"b4":"b4v",
					"b5":"b5v"
				}
			]`,
		depth: 5,
		expectedKeys: []string{
			"b1", "b2", "b3", "b4", "b5",
		},
	})
}

func TestErrors(t *testing.T) {
	testDocCase(t, docTest{
		doc: `{
			"a":{
				"a1": {
					"a11":1
				},
				"a2":"a2v"
			}
		`,
		depth:        5,
		expectedKeys: nil,
		expectedErr:  "unterminated JSON object",
	})
	testDocCase(t, docTest{
		doc: `[
				{
					"b1":"b1v"
				},
				{
					"b2":"b2v"
				},
				{
					"b3":"b3v"
				},
				{
					"b4":"b4v",
					"b5":"b5v"
				}`,
		depth:        5,
		expectedKeys: nil,
		expectedErr:  "unterminated JSON array",
	})
	testDocCase(t, docTest{
		doc:          `"hello"`,
		depth:        5,
		expectedKeys: nil,
		expectedErr:  "sample is not an object or array",
	})
	testDocCase(t, docTest{
		doc: `{
			"a":{
				"a1": {
					"a11":1
				},
				"a2":"a2v"
				{}
			}
		}`,
		depth:        5,
		expectedKeys: nil,
		expectedErr:  "Value looks like object, but can't find closing '}' symbol",
	})
	testDocCase(t, docTest{
		doc: `{
			"a":^
		}`,
		depth:        5,
		expectedKeys: nil,
		expectedErr:  "Unknown value type",
	})
	testDocCase(t, docTest{
		doc: `{
			"a":[1,^,2]
		}`,
		depth:        5,
		expectedKeys: nil,
		expectedErr:  "Unknown value type",
	})
	testDocCase(t, docTest{
		doc: `{
			"a":[
				1,
				{
					"a1":tru
				},
				3
			]
		}`,
		depth:        5,
		expectedKeys: nil,
		expectedErr:  "Unknown value type",
	})
}

func testDocCase(t *testing.T, tCase docTest) {
	t.Helper()
	defer func() {
		if tpanic := recover(); tpanic != nil {
			t.Fatalf("test panic: %v", tpanic)
		}
	}()
	ks := New(5)
	// sample doc once
	err := ks.Sample([]byte(tCase.doc))
	// get all keys without frequency
	keys := ks.Keys()
	sort.Strings(keys)
	sort.Strings(tCase.expectedKeys)
	if len(keys) != len(tCase.expectedKeys) {
		t.Fatalf("expected %d keys - got: %d", len(tCase.expectedKeys), len(keys))
	}
	for i, k := range tCase.expectedKeys {
		if keys[i] != k {
			t.Fatalf("key %d/%d did not match\nexpected: %s\ngot:%s", i+1, len(keys)+1, tCase.expectedKeys[i], keys[i])
		}
	}

	if len(tCase.expectedErr) > 0 {
		if err == nil {
			t.Logf("expected error %q - got nothing", tCase.expectedErr)
			t.Fail()
		}
		if !strings.HasPrefix(err.Error(), tCase.expectedErr) {
			t.Logf("expected error to match %q - got %q", tCase.expectedErr, err.Error())
			t.Fail()
		}
	} else {
		if err != nil {
			t.Logf("expected no errors - got %q", err.Error())
			t.Fail()
		}
	}
}
