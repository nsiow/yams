package policy

import (
	"encoding/json"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

// TestNewValue creates a Value with different variables and determines correct functionality
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

// TestMarshal validates the JSON marshalling behavior of various cases
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

// TestUnmarshal validates the JSON unmarshalling behavior of various cases
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
		{Input: `{"S": "\""}`, ShouldErr: true},
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

// TestInvalid validates the handling of invalid JSON fragments
func TestInvalid(t *testing.T) {
	tests := []testlib.TestCase[string, any]{
		{Input: `a`, ShouldErr: true},
	}

	testlib.RunTestSuite(t, tests, func(s string) (any, error) {
		var v Value
		return nil, json.Unmarshal([]byte(s), &v)
	})
}

// TestEmpty validates the correct emptiness behavior of a Value
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
