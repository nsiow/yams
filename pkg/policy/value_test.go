package policy

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestNewValue creates a Value with different variables and determines correct functionality
func TestNewValue(t *testing.T) {
	type test struct {
		input Value
		want  Value
	}

	tests := []test{
		{input: []string{}, want: []string{}},
	}

	for _, tc := range tests {
		got := NewValue(tc.input...)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %#v, got: %#v", tc.want, got)
		}
	}
}

// TestUnmarshal validates the JSON unmarshalling behavior of various cases
func TestUnmarshalValid(t *testing.T) {
	type exampleStruct struct {
		S Value
	}

	type test struct {
		input string
		want  Value
		err   bool
	}

	tests := []test{
		{input: `{"S": "foo"}`, want: []string{"foo"}},
		{input: `{"S": ["foo", "bar"]}`, want: []string{"foo", "bar"}},
		{input: `{"S": ""}`, want: []string{""}},
		{input: `{"S": null}`, want: []string{}},
		{input: `{"S": "null"}`, want: []string{"null"}},
		{input: `{"S": []}`, want: []string{}},
		{input: `{"S": [0]}`, err: true},
		{input: `{"S": 0}`, err: true},
		{input: `{"S": true`, err: true},
	}

	for _, tc := range tests {
		ex := exampleStruct{}
		err := json.Unmarshal([]byte(tc.input), &ex)
		if tc.err {
			if err == nil {
				t.Fatalf("expected: err, got: nil, for input: %#v", tc.input)
			} else {
				continue // expected error, saw error; next test
			}
		}
		if err != nil {
			t.Fatalf("error unmarshalling Value: %v", err)
		}

		got := ex.S
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %#v, got: %#v, for input: %#v", tc.want, got, tc.input)
		}
	}
}

// TestInvalid validates the handling of invalid JSON fragments
func TestInvalid(t *testing.T) {
	type test struct {
		input string
	}

	tests := []test{
		{input: `a`},
	}

	for _, tc := range tests {
		var v Value
		err := json.Unmarshal([]byte(tc.input), &v)
		if err != nil {
			t.Logf("test saw expected error: %v", err)
			continue
		}

		t.Fatalf("expected error, got success for input: %s", tc.input)
	}
}

// TestEmpty validates the correct emptiness behavior of a Value
func TestEmpty(t *testing.T) {
	type test struct {
		input Value
		want  bool
	}

	tests := []test{
		{input: nil, want: true},
		{input: []string{}, want: true},
		{input: []string{"foo"}, want: false},
		{input: []string{"foo", "bar"}, want: false},
		{input: []string{"foo", "bar", "baz"}, want: false},
	}

	for _, tc := range tests {
		got := tc.input.Empty()
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %#v, got: %#v, for input: %#v", tc.want, got, tc.input)
		}
	}
}
