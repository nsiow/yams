package awsconfig

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/policy"
)

// Testing a handful of edge cases that cannot be reached from elsewhere in the code

// TestExtractInlinePoliciesInvalid confirms correct error handling behavior for unexpected types
func TestExtractInlinePoliciesInvalid(t *testing.T) {
	tests := []testlib.TestCase[ConfigItem, []policy.Policy]{
		{
			Input: ConfigItem{
				Type: "AWS::IAM::Policy",
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, extractInlinePolicies)
}

// TestExtractManagedPoliciesInvalid confirms correct error handling behavior for unexpected types
func TestExtractManagedPoliciesInvalid(t *testing.T) {
	tests := []testlib.TestCase[ConfigItem, []policy.Policy]{
		{
			Input: ConfigItem{
				Type: "AWS::IAM::Policy",
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(c ConfigItem) ([]policy.Policy, error) {
		return extractManagedPolicies(c, nil)
	})
}

// TestExtractPermissionsBoundaryInvalid confirms correct error handling behavior for unexpected types
func TestExtractPermissionsBoundaryInvalid(t *testing.T) {
	tests := []testlib.TestCase[ConfigItem, policy.Policy]{
		{
			Input: ConfigItem{
				Type: "AWS::IAM::Policy",
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(c ConfigItem) (policy.Policy, error) {
		return extractPermissionsBoundary(c, nil)
	})
}
