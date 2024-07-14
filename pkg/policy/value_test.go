package policy

import (
	"encoding/json"
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
)

// TestNewValue creates a Value with different variables and determines correct functionality
func TestNewValue(t *testing.T) {
	tests := []testrunner.TestCase[[]string, []string]{
		{
			Input: []string{},
			Want:  []string{},
		},
	}

	testrunner.RunTestSuite(t, tests, func(v []string) ([]string, error) {
		got := NewValue(v...)
		return got, nil
	})
}

// TestUnmarshal validates the JSON unmarshalling behavior of various cases
func TestUnmarshalValid(t *testing.T) {
	type exampleStruct struct {
		S Value
	}

	tests := []testrunner.TestCase[string, Value]{
		{Input: `{"S": "foo"}`, Want: []string{"foo"}},
		{Input: `{"S": ["foo", "bar"]}`, Want: []string{"foo", "bar"}},
		{Input: `{"S": null}`, Want: []string{}},
		{Input: `{"S": "null"}`, Want: []string{"null"}},
		{Input: `{"S": []}`, Want: []string{}},
		{Input: `{"S": true}`, Want: []string{"true"}},
		{Input: `{"S": false}`, Want: []string{"false"}},
		{Input: `{"S": ""}`, ShouldErr: true},
		{Input: `{"S": [0]}`, ShouldErr: true},
		{Input: `{"S": 0}`, ShouldErr: true},
		{Input: `{"S": 1000}`, ShouldErr: true},
		{Input: `{"S": true`, ShouldErr: true},
	}

	testrunner.RunTestSuite(t, tests, func(s string) (Value, error) {
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
	tests := []testrunner.TestCase[string, any]{
		{Input: `a`, ShouldErr: true},
	}

	testrunner.RunTestSuite(t, tests, func(s string) (any, error) {
		var v Value
		return nil, json.Unmarshal([]byte(s), &v)
	})
}

// TestEmpty validates the correct emptiness behavior of a Value
func TestEmpty(t *testing.T) {
	tests := []testrunner.TestCase[Value, bool]{
		{Input: nil, Want: true},
		{Input: []string{}, Want: true},
		{Input: []string{"foo"}, Want: false},
		{Input: []string{"foo", "bar"}, Want: false},
		{Input: []string{"foo", "bar", "baz"}, Want: false},
	}

	testrunner.RunTestSuite(t, tests, func(v Value) (bool, error) {
		return v.Empty(), nil
	})
}
