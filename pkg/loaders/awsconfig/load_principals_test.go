package awsconfig

import (
	"testing"
)

// Testing a handful of edge cases that cannot be reached from elsewhere in the code

// TestExtractInlinePoliciesInvalid confirms correct error handling behavior for unexpected types
func TestExtractInlinePoliciesInvalid(t *testing.T) {
	type test struct {
		input ConfigItem
	}

	tests := []test{
		{
			input: ConfigItem{
				Type: "AWS::IAM::Policy",
			},
		},
	}

	for _, tc := range tests {
		_, err := extractInlinePolicies(tc.input)
		if err == nil {
			t.Fatalf("expected error but saw success for input: %#v", tc.input)
		}

		t.Logf("saw expected error: %v", err)
	}
}

// TestExtractManagedPoliciesInvalid confirms correct error handling behavior for unexpected types
func TestExtractManagedPoliciesInvalid(t *testing.T) {
	type test struct {
		input ConfigItem
	}

	tests := []test{
		{
			input: ConfigItem{
				Type: "AWS::IAM::Policy",
			},
		},
	}

	for _, tc := range tests {
		_, err := extractManagedPolicies(tc.input, nil)
		if err == nil {
			t.Fatalf("expected error but saw success for input: %#v", tc.input)
		}

		t.Logf("saw expected error: %v", err)
	}
}
