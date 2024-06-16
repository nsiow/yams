package awsconfig

import (
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
	"github.com/nsiow/yams/pkg/policy"
)

// TestPolicyStorageRetrieval confirms the ability to store/retrieve managed policies correctly
func TestPolicyStorageRetrieval(t *testing.T) {
	type input struct {
		store    bool            // whether or not to store the provided policy before the test
		arn      string          // the ARN to use for storage + retrieval
		policies []policy.Policy // the policy to store
	}

	type output struct {
		exists   bool            // whether or not we should expect the requested ARN to exist
		policies []policy.Policy // the policy we would expect to get back
	}

	// Define inputs
	tests := []testrunner.TestCase[input, output]{
		{
			Name: "empty_policy",
			Input: input{
				store:    true,
				arn:      "arn:aws:iam::000000000000:policy/EmptyPolicy",
				policies: nil,
			},
			Want: output{
				exists:   true,
				policies: nil,
			},
		},
		{
			Name: "s3read_basic",
			Input: input{
				store: true,
				arn:   "arn:aws:iam::000000000000:policy/AmazonFakeS3ReadOnlyAccess",
				policies: []policy.Policy{
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
			Want: output{
				exists: true,
				policies: []policy.Policy{
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
		{
			Name: "normalize_1",
			Input: input{
				store: true,
				arn:   "arn:aws:iam::aws:policy/NormalizationTest",
				policies: []policy.Policy{
					{
						Id: "something_unique",
					},
				},
			},
			Want: output{
				exists: true,
				policies: []policy.Policy{
					{
						Id: "something_unique",
					},
				},
			},
		},
		{
			Name: "normalize_2",
			Input: input{
				store: true,
				arn:   "arn:aws:iam::aws:policy/aws-service-role/NormalizationTest",
				policies: []policy.Policy{
					{
						Id: "something_unique",
					},
				},
			},
			Want: output{
				exists: true,
				policies: []policy.Policy{
					{
						Id: "something_unique",
					},
				},
			},
		},
		{
			Name: "normalize_3",
			Input: input{
				store: true,
				arn:   "arn:aws:iam::aws:policy/service-role/NormalizationTest",
				policies: []policy.Policy{
					{
						Id: "something_unique",
					},
				},
			},
			Want: output{
				exists: true,
				policies: []policy.Policy{
					{
						Id: "something_unique",
					},
				},
			},
		},
		{
			Name: "should_be_missing",
			Input: input{
				store: false,
				arn:   "arn:aws:iam::aws:policy/NonexistentPolicy",
			},
			Want: output{
				exists:   false,
				policies: nil,
			},
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (output, error) {
		// Create a new policy map
		m := NewPolicyMap()

		// Store the policy if requested
		if i.store {
			m.Add(CONST_TYPE_AWS_IAM_POLICY, i.arn, i.policies)
		}

		// Retrieve + return the policy, formatting into `output`
		policies, exists := m.Get(CONST_TYPE_AWS_IAM_POLICY, i.arn)
		got := output{exists: exists, policies: policies}
		return got, nil
	})
}
