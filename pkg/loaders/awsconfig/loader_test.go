package awsconfig

import (
	"os"
	"reflect"
	"testing"

	"github.com/nsiow/yams/pkg/entities"
)

// Define some common test variables here, which we'll use across multiple tests
type output struct {
	Principals []entities.Principal
	Resources  []entities.Resource
	Policies   *PolicyMap
}

var simple1Output output = output{}

// TestLoadJson confirms that we can correctly load data from JSON arrays of AWS Config data
func TestLoadJsonValid(t *testing.T) {
	type test struct {
		name  string
		input string
		want  output
	}

	tests := []test{
		{
			name:  "empty",
			input: `../../../testdata/environments/empty.json`,
			want: output{
				Principals: []entities.Principal(nil),
				Resources:  []entities.Resource(nil),
				Policies:   nil,
			},
		},
		{
			name:  "simple_1",
			input: `../../../testdata/environments/simple_1.json`,
			want:  simple1Output,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		inputBytes, err := os.ReadFile(tc.input)
		if err != nil {
			t.Fatalf("unable to read file '%s' for test case: '%s': %v", tc.input, tc.name, err)
		}

		l := NewLoader()
		err = l.LoadJson(inputBytes)
		if err != nil {
			t.Fatalf("unexpected error for test case: '%s': %v", tc.name, err)
		}

		got := output{
			Principals: l.Principals(),
			Resources:  l.Resources(),
		}

		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %#v, got: %#v, for test case '%s'", tc.want, got, tc.name)
		}
	}
}
