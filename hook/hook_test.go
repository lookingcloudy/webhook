package hook

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

var hooksLoadFromFileTests = []struct {
	path string
	ok   bool
}{
	{"testdata/hooks-branch.json", true},
	{"testdata/hooks-multi.json", true},
	{"", true},
	// failures
	{"missing.json", false},
}

func TestHooksLoadFromFile(t *testing.T) {
	for _, tt := range hooksLoadFromFileTests {
		h := &Hooks{}
		err := h.LoadFromFile(tt.path)
		if (err == nil) != tt.ok {
			t.Errorf(err.Error())
		}
	}
}

var hooksMatchTests = []struct {
	id    string
	hooks Hooks
	value *Hook
}{
	{"a", Hooks{Hook{ID: "a"}}, &Hook{ID: "a"}},
	{"X", Hooks{Hook{ID: "a"}}, new(Hook)},
}

func TestHooksMatch(t *testing.T) {
	for _, tt := range hooksMatchTests {
		value := tt.hooks.Match(tt.id)
		if reflect.DeepEqual(reflect.ValueOf(value), reflect.ValueOf(tt.value)) {
			t.Errorf("failed to match %q:\nexpected %#v,\ngot %#v", tt.id, tt.value, value)
		}
	}
}

var bitbucketTest = []struct {
	path     string
	testType string
	expected []string
}{
	{"testdata/bitbucket-branch.json", "branch", []string{"develop"}},
	{"testdata/bitbucket-tag.json", "tag", []string{"v10.1-qa"}},
}

func TestBitbucket(t *testing.T) {
	for _, tt := range bitbucketTest {
		// parse hook file for hooks
		file, e := ioutil.ReadFile(tt.path)

		if e != nil {
			t.Errorf("Could not read file %s\n", tt.path)
			return
		}
		b := BitPush{}
		e = json.Unmarshal(file, &b)
		switch tt.testType {
		case "branch":

			if !reflect.DeepEqual(b.GetBranches(), tt.expected) {
				//if !compareSlices(b.GetBranches(), tt.expected) {
				t.Errorf("Test: %s, Expected %v, got %v\n", tt.path, tt.expected, b.GetBranches())
			}
		case "tag":
			if !reflect.DeepEqual(b.GetTags(), tt.expected) {
				t.Errorf("Test: %s, Expected %v, got %v\n", tt.path, tt.expected, b.GetBranches())
			}
		}
	}
}

var evaluateTests = []struct {
	hookFile    string
	match       string
	webhookFile string
	expected    string
}{
	{"testdata/hooks-branch.json", "develop", "testdata/bitbucket-branch.json", "develop"},
	{"testdata/hooks-branch.json", "develop", "testdata/bitbucket-tag.json", ""},
	{"testdata/hooks-tag.json", "qa-builds", "testdata/bitbucket-tag.json", "v10.1-qa"},
}

func TestEvaluate(t *testing.T) {
	for _, tt := range evaluateTests {
		b := getBitBucket(t, tt.webhookFile)
		h := getHooks(t, tt.hookFile)
		_, match := h.Match(tt.match).Evaluate(b)
		if tt.expected != match {
			t.Errorf("Hook %s, bitbucket %s, expected %s, got %s\n", tt.hookFile, tt.webhookFile, tt.expected, match)
		}
	}
}

func getHooks(t *testing.T, f string) Hooks {
	h := Hooks{}
	e := h.LoadFromFile(f)
	if e != nil {
		t.Errorf("Error loading hooks from file %f: %v\n", f, e)
	}
	return h
}

func getBitBucket(t *testing.T, f string) *BitPush {
	b := BitPush{}
	e := getStructFromFile(f, &b)
	if e != nil {
		t.Errorf("Error loading BitPush from file %f: %v\n", f, e)
	}
	return &b
}

func getStructFromFile(f string, s interface{}) error {
	file, e := ioutil.ReadFile(f)

	if e != nil {
		return e
	}
	e = json.Unmarshal(file, s)

	if e != nil {
		return e
	}
	return nil
}
