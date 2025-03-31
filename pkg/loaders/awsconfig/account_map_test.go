package awsconfig

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// TestAccountStorageRetrieval confirms the ability to store/retrieve accounts directly
func TestAccountStorageRetrieval(t *testing.T) {
	type input struct {
		store     bool             // whether or not to store the provided account before the test
		accountId string           // the accountId to use for storage + retrieval
		account   entities.Account // the account to store
	}

	type output struct {
		exists  bool             // whether or not we should expect the requested account to exist
		account entities.Account // the account we would expect to get back
	}

	// Define inputs
	tests := []testlib.TestCase[input, output]{
		{
			Name: "empty_policy",
			Input: input{
				store:     true,
				accountId: "000000000000",
				account:   entities.Account{},
			},
			Want: output{
				exists:  true,
				account: entities.Account{},
			},
		},
		{
			Name: "multiple_levels",
			Input: input{
				store:     true,
				accountId: "000000000000",
				account: entities.Account{
					Id:       "000000000000",
					OrgId:    "o-123",
					OrgPaths: []string{"o-123/", "o-123/ou-level-1/", "o-123/ou-level-1/ou-level-2/"},
					SCPs: [][]policy.Policy{
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
			Want: output{
				exists: true,
				account: entities.Account{
					Id:       "000000000000",
					OrgId:    "o-123",
					OrgPaths: []string{"o-123/", "o-123/ou-level-1/", "o-123/ou-level-1/ou-level-2/"},
					SCPs: [][]policy.Policy{
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
		},
		{
			Name: "s3read_basic",
			Input: input{
				store: true,
				account: entities.Account{
					Id:       "000000000000",
					OrgId:    "o-123",
					OrgPaths: []string{"o-123/", "o-123/ou-level-1/", "o-123/ou-level-1/ou-level-2/"},
					SCPs: [][]policy.Policy{
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
			Want: output{
				exists: true,
				account: entities.Account{
					Id:       "000000000000",
					OrgId:    "o-123",
					OrgPaths: []string{"o-123/", "o-123/ou-level-1/", "o-123/ou-level-1/ou-level-2/"},
					SCPs: [][]policy.Policy{
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
		},
		{
			Name: "should_be_missing",
			Input: input{
				store: false,
			},
			Want: output{
				exists: false,
			},
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (output, error) {
		// Create a new account map
		m := NewAccountMap()

		// Store the policy if requested
		if i.store {
			m.Add(i.accountId, i.account)
		}

		// Retrieve + return the account, formatting into `output`
		account, exists := m.Get(i.accountId)
		got := output{exists: exists, account: account}
		return got, nil
	})
}
