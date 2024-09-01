package awsconfig

import (
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
	"github.com/nsiow/yams/pkg/policy"
)

// TestControlPolicyStorageRetrieval confirms the ability to store/retrieve control policies
// correctly
func TestControlPolicyStorageRetrieval(t *testing.T) {
	type input struct {
		store    bool              // whether or not to store the provided policy before the test
		account  string            // the ARN to use for storage + retrieval
		policies [][]policy.Policy // the policy to store
	}

	type output struct {
		exists   bool              // whether or not we should expect the requested ARN to exist
		policies [][]policy.Policy // the policy we would expect to get back
	}

	// Define inputs
	tests := []testrunner.TestCase[input, output]{
		{
			Name: "empty_policy",
			Input: input{
				store:    true,
				account:  "000000000000",
				policies: nil,
			},
			Want: output{
				exists:   true,
				policies: nil,
			},
		},
		{
			Name: "multiple_levels",
			Input: input{
				store:   true,
				account: "000000000000",
				policies: [][]policy.Policy{
					{
						{
							Version: "2012-10-17",
							Id:      "allowAll",
							Statement: []policy.Statement{
								{
									Effect: "Allow",
									Action: []string{
										"*:*",
									},
									Resource: []string{
										"*",
									},
								},
							},
						},
						{
							Version: "2012-10-17",
							Id:      "allowS3",
							Statement: []policy.Statement{
								{
									Effect: "Allow",
									Action: []string{
										"s3:*",
										"s3:ListBucket",
									},
									Resource: []string{
										"*",
									},
								},
							},
						},
					},
				},
			},
			Want: output{
				exists: true,
				policies: [][]policy.Policy{
					{
						{
							Version: "2012-10-17",
							Id:      "allowAll",
							Statement: []policy.Statement{
								{
									Effect: "Allow",
									Action: []string{
										"*:*",
									},
									Resource: []string{
										"*",
									},
								},
							},
						},
						{
							Version: "2012-10-17",
							Id:      "allowS3",
							Statement: []policy.Statement{
								{
									Effect: "Allow",
									Action: []string{
										"s3:*",
										"s3:ListBucket",
									},
									Resource: []string{
										"*",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "s3read_basic",
			Input: input{
				store:   true,
				account: "000000000000",
				policies: [][]policy.Policy{
					{
						{
							Version: "2012-10-17",
							Id:      "s3read",
							Statement: []policy.Statement{
								{
									Effect: "Allow",
									Action: []string{
										"s3:GetObject",
										"s3:ListBucket",
									},
									Resource: []string{
										"arn:aws:s3:::foo-bucket",
										"arn:aws:s3:::foo-bucket/*",
									},
								},
							},
						},
					},
				},
			},
			Want: output{
				exists: true,
				policies: [][]policy.Policy{
					{
						{
							Version: "2012-10-17",
							Id:      "s3read",
							Statement: []policy.Statement{
								{
									Effect: "Allow",
									Action: []string{
										"s3:GetObject",
										"s3:ListBucket",
									},
									Resource: []string{
										"arn:aws:s3:::foo-bucket",
										"arn:aws:s3:::foo-bucket/*",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "should_be_missing",
			Input: input{
				store:   false,
				account: "000000000000",
			},
			Want: output{
				exists:   false,
				policies: nil,
			},
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (output, error) {
		// Create a new policy map
		m := NewControlPolicyMap()

		// Store the policy if requested
		if i.store {
			m.Add(i.account, i.policies)
		}

		// Retrieve + return the policy, formatting into `output`
		policies, exists := m.Get(i.account)
		got := output{exists: exists, policies: policies}
		return got, nil
	})
}
