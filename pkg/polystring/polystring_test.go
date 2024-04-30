package polystring

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestNewPolyString creates a polystring with different variables and determines correct functionality
func TestNewPolyString(t *testing.T) {
	type test struct {
		input []string
		want  []string
	}

	tests := []test{
		{input: []string{}, want: []string{}},
	}

	for _, tc := range tests {
		got := NewPolyString(tc.input...).Values
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}

// TestUnmarshal validates the JSON unmarshalling behavior of various cases
func TestUnmarshalValid(t *testing.T) {
	type exampleStruct struct {
		S PolyString
	}

	type test struct {
		input string
		want  []string
		err   bool
	}

	tests := []test{
		{input: `{"S": "foo"}`, want: []string{"foo"}},
		{input: `{"S": ["foo", "bar"]}`, want: []string{"foo", "bar"}},
		{input: `{"S": ""}`, want: []string{""}},
		{input: `{"S": null}`, want: nil},
		{input: `{"S": "null"}`, want: []string{"null"}},
		{input: `{"S": []}`, want: []string{}},
		{input: `{"S": [0]}`, err: true},
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
			t.Fatalf("error unmarshalling polystring: %v", err)
		}

		got := ex.S.Values
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %#v, got: %#v, for input: %#v", tc.want, got, tc.input)
		}
	}
}

// TestEmpty validates the correct emptiness behavior of a PolyString
func TestEmpty(t *testing.T) {
	type test struct {
		input PolyString
		want  bool
	}

	tests := []test{
		{input: PolyString{Values: nil}, want: true},
		{input: PolyString{Values: []string{}}, want: true},
		{input: PolyString{Values: []string{"foo"}}, want: false},
		{input: PolyString{Values: []string{"foo", "bar"}}, want: false},
		{input: PolyString{Values: []string{"foo", "bar", "baz"}}, want: false},
	}

	for _, tc := range tests {
		got := tc.input.Empty()
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %#v, got: %#v, for input: %#v", tc.want, got, tc.input)
		}
	}
}
