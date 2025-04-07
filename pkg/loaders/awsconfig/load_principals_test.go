package awsconfig

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/policy"
)

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
