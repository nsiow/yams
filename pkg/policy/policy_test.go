package policy

import (
	"encoding/json"
	"reflect"
	"testing"

	tu "github.com/nsiow/yams/internal/testutil"
)

// TestPolicyGrammar runs a series of tests to confirm we properly parse the IAM policy grammar
func TestPolicyGrammar(t *testing.T) {
	type test struct {
		asset string
		want  Policy
	}

	tests := []test{
		{
			asset: "empty_policy",
			want: Policy{
				Version:   "",
				Id:        "",
				Statement: StatementBlock{Values: []Statement{}},
			},
		},
	}

	for _, tc := range tests {
		// Read asset from file
		asset, err := tu.ReadBytesFromTestAsset(tc.asset)
		if err != nil {
			t.Fatalf("unable to load test asset '%s': %v", tc.asset, err)
		}

		// Unmarshal into a policy
		p := Policy{}
		err = json.Unmarshal(asset, &p)
		if err != nil {
			t.Fatalf("unable to create policy from asset '%s': %v", tc.asset, err)
		}

		// Check against expected value
		if !reflect.DeepEqual(tc.want, p) {
			t.Fatalf("expected: %#v, got: %#v", tc.want, p)
		}
	}
}
