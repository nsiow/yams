package awsconfig

import (
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
	"github.com/nsiow/yams/pkg/policy"
)

// Testing a handful of edge cases that cannot be reached from elsewhere in the code

// TestExtractInlinePoliciesInvalid confirms correct error handling behavior for unexpected types
func TestExtractInlinePoliciesInvalid(t *testing.T) {
	tests := []testrunner.TestCase[ConfigItem, []policy.Policy]{
		{
			Input: ConfigItem{
				Type: "AWS::IAM::Policy",
			},
			ShouldErr: true,
		},
	}

	testrunner.RunTestSuite(t, tests, extractInlinePolicies)
}

// TestExtractManagedPoliciesInvalid confirms correct error handling behavior for unexpected types
func TestExtractManagedPoliciesInvalid(t *testing.T) {
	tests := []testrunner.TestCase[ConfigItem, []policy.Policy]{
		{
			Input: ConfigItem{
				Type: "AWS::IAM::Policy",
			},
			ShouldErr: true,
		},
	}

	testrunner.RunTestSuite(t, tests, func(c ConfigItem) ([]policy.Policy, error) {
		return extractManagedPolicies(c, nil)
	})
}
