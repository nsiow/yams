package policy

import (
	"encoding/json"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestNewValue(t *testing.T) {
	tests := []testlib.TestCase[Value, []string]{
		{
			Input: Value{},
			Want:  Value{},
		},
	}

	testlib.RunTestSuite(t, tests, func(v Value) ([]string, error) {
		got := NewValue(v...)
		return got, nil
	})
}

func TestMarshal(t *testing.T) {
	type exampleStruct struct {
		S Value
	}

	tests := []testlib.TestCase[Value, string]{
		{Input: Value{"foo"}, Want: `{"S":"foo"}`},
		// {Input: Value{"foo", "bar"}, Want: `{"S":["foo","bar"]}`},
		// {Input: Value{}, Want: `{"S":[]}`},
		// {Input: Value{"null"}, Want: `{"S":["null"]}`},
		// {Input: Value{}, Want: `{"S":[]}`},
		// {Input: Value{"true"}, Want: `{"S":"true"}`},
		// {Input: Value{"false"}, Want: `{"S":"false"}`},
	}

	testlib.RunTestSuite(t, tests, func(s Value) (string, error) {
		ex := exampleStruct{S: s}
		b, err := json.Marshal(ex)
		if err != nil {
			return "", err
		}

		return string(b), nil
	})
}

func TestUnmarshal(t *testing.T) {
	type exampleStruct struct {
		S Value
	}

	tests := []testlib.TestCase[string, Value]{
		{Input: `{"S": "foo"}`, Want: Value{"foo"}},
		{Input: `{"S": ["foo", "bar"]}`, Want: Value{"foo", "bar"}},
		{Input: `{"S": null}`, Want: Value{}},
		{Input: `{"S": "null"}`, Want: Value{"null"}},
		{Input: `{"S": []}`, Want: Value{}},
		{Input: `{"S": true}`, Want: Value{"true"}},
		{Input: `{"S": false}`, Want: Value{"false"}},
		{Input: `{"S": "\""}`, Want: Value{`"`}},
		{Input: `{"S": [0]}`, ShouldErr: true},
		{Input: `{"S": 0}`, ShouldErr: true},
		{Input: `{"S": 1000}`, ShouldErr: true},
		{Input: `{"S": true`, ShouldErr: true},
	}

	testlib.RunTestSuite(t, tests, func(s string) (Value, error) {
		ex := exampleStruct{}
		err := json.Unmarshal([]byte(s), &ex)
		if err != nil {
			return Value{}, err
		}

		return ex.S, nil
	})
}

func TestInvalid(t *testing.T) {
	tests := []testlib.TestCase[string, any]{
		{Input: `a`, ShouldErr: true},
	}

	testlib.RunTestSuite(t, tests, func(s string) (any, error) {
		var v Value
		return nil, json.Unmarshal([]byte(s), &v)
	})
}

func TestValue_UnmarshalJSON_ErrorPaths(t *testing.T) {
	// Test error path in single-value clause (line 56-57)
	// A string that starts with " but isn't valid JSON
	var v1 Value
	err := v1.UnmarshalJSON([]byte(`"unterminated`))
	if err == nil {
		t.Fatal("expected error for invalid single-value JSON but got nil")
	}

	// Test error path in multi-value clause (line 66-67)
	// An array that starts with [ but contains invalid elements
	var v2 Value
	err = v2.UnmarshalJSON([]byte(`[1, 2, 3]`))
	if err == nil {
		t.Fatal("expected error for invalid array elements but got nil")
	}
}

func TestEmpty(t *testing.T) {
	tests := []testlib.TestCase[Value, bool]{
		{Input: nil, Want: true},
		{Input: Value{}, Want: true},
		{Input: Value{"foo"}, Want: false},
		{Input: Value{"foo", "bar"}, Want: false},
		{Input: Value{"foo", "bar", "baz"}, Want: false},
	}

	testlib.RunTestSuite(t, tests, func(v Value) (bool, error) {
		return v.Empty(), nil
	})
}

func TestContains(t *testing.T) {
	type input struct {
		haystack Value
		needle   string
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_contains",
			Input: input{
				haystack: Value{"red", "green", "blue"},
				needle:   "red",
			},
			Want: true,
		},
		{
			Name: "empty_value",
			Input: input{
				haystack: Value{},
				needle:   "red",
			},
			Want: false,
		},
		{
			Name: "respect_case",
			Input: input{
				haystack: Value{"quick", "brown", "fox"},
				needle:   "FOX",
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		return i.haystack.Contains(i.needle), nil
	})
}
