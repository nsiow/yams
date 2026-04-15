package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestIsStrictCall(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, bool]{
		{
			Name:  "empty_subject",
			Input: AuthContext{},
			Want:  false,
		},
		{
			Name: "non_strict_call",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:getobject"),
			},
			Want: false,
		},
		{
			Name: "sts_assume_role",
			Input: AuthContext{
				Action:    sar.MustLookupString("sts:assumerole"),
				Principal: &entities.FrozenPrincipal{},
				Resource:  &entities.FrozenResource{},
			},
			Want: true,
		},
		{
			Name: "kms_plus_key",
			Input: AuthContext{
				Action:    sar.MustLookupString("kms:decrypt"),
				Principal: &entities.FrozenPrincipal{},
				Resource: &entities.FrozenResource{
					Arn:  "arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab",
					Type: "AWS::KMS::Key",
				},
			},
			Want: true,
		},
		{
			Name: "kms_sans_key",
			Input: AuthContext{
				Action:    sar.MustLookupString("kms:decrypt"),
				Principal: &entities.FrozenPrincipal{},
				Resource: &entities.FrozenResource{
					Arn:  "arn:aws:kms:us-west-2:111122223333:alias/ExampleAlias",
					Type: "AWS::KMS::Alias",
				},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i AuthContext) (bool, error) {
		subj := newSubject(i, TestingSimulationOptions)
		return isStrictCall(&subj), nil
	})
}

func TestResourceAccessGrantsPrincipal(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, bool]{
		// Direct grants (should return true)
		{
			Name: "grant_explicit_arn",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Type:      "AWS::IAM::Role",
					Arn:       "arn:aws:iam::55555:role/MyRole",
					AccountId: "55555",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::nsiow-test",
					Type:      "AWS::S3::Bucket",
					AccountId: "55555",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Sid: "test_statement",
								Principal: policy.Principal{
									AWS: policy.Value{"arn:aws:iam::55555:role/MyRole"},
								},
								Effect: "Allow",
								Action: []string{"s3:listbucket"},
							},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "grant_principal_star_in_aws",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Type:      "AWS::IAM::Role",
					Arn:       "arn:aws:iam::55555:role/MyRole",
					AccountId: "55555",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::nsiow-test",
					Type:      "AWS::S3::Bucket",
					AccountId: "55555",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Sid: "test_statement",
								Principal: policy.Principal{
									AWS: policy.Value{"*"},
								},
								Effect: "Allow",
								Action: []string{"s3:listbucket"},
							},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "grant_principal_all",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Type:      "AWS::IAM::Role",
					Arn:       "arn:aws:iam::55555:role/MyRole",
					AccountId: "55555",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::nsiow-test",
					Type:      "AWS::S3::Bucket",
					AccountId: "55555",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Sid:       "test_statement",
								Principal: policy.Principal{All: true},
								Effect:    "Allow",
								Action:    []string{"s3:listbucket"},
							},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "grant_mixed_delegation_and_arn",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Type:      "AWS::IAM::Role",
					Arn:       "arn:aws:iam::55555:role/MyRole",
					AccountId: "55555",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::nsiow-test",
					Type:      "AWS::S3::Bucket",
					AccountId: "55555",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Sid: "test_statement",
								Principal: policy.Principal{
									AWS: policy.Value{"55555", "arn:aws:iam::55555:role/MyRole"},
								},
								Effect: "Allow",
								Action: []string{"s3:listbucket"},
							},
						},
					},
				},
			},
			Want: true,
		},

		// Delegated access (should return false)
		{
			Name: "delegated_account_id",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Type:      "AWS::IAM::Role",
					Arn:       "arn:aws:iam::55555:role/MyRole",
					AccountId: "55555",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::nsiow-test",
					Type:      "AWS::S3::Bucket",
					AccountId: "55555",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Sid: "test_statement",
								Principal: policy.Principal{
									AWS: policy.Value{"55555"},
								},
								Effect: "Allow",
								Action: []string{"s3:listbucket"},
							},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "delegated_account_root",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Type:      "AWS::IAM::Role",
					Arn:       "arn:aws:iam::55555:role/MyRole",
					AccountId: "55555",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::nsiow-test",
					Type:      "AWS::S3::Bucket",
					AccountId: "55555",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Sid: "test_statement",
								Principal: policy.Principal{
									AWS: policy.Value{"arn:aws:iam::55555:root"},
								},
								Effect: "Allow",
								Action: []string{"s3:listbucket"},
							},
						},
					},
				},
			},
			Want: false,
		},

		// Non-matching cases (should return false)
		{
			Name: "unrelated_action",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Type:      "AWS::IAM::Role",
					Arn:       "arn:aws:iam::55555:role/MyRole",
					AccountId: "55555",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::nsiow-test",
					Type:      "AWS::S3::Bucket",
					AccountId: "55555",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Sid: "test_statement",
								Principal: policy.Principal{
									AWS: policy.Value{"arn:aws:iam::55555:role/MyRole"},
								},
								Effect: "Allow",
								Action: []string{"s3:getbucketpolicy"},
							},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "principal_star_unrelated_action",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Type:      "AWS::IAM::Role",
					Arn:       "arn:aws:iam::55555:role/MyRole",
					AccountId: "55555",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::nsiow-test",
					Type:      "AWS::S3::Bucket",
					AccountId: "55555",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Sid: "test_statement",
								Principal: policy.Principal{
									AWS: policy.Value{"*"},
								},
								Effect: "Allow",
								Action: []string{"s3:getobject"},
							},
						},
					},
				},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i AuthContext) (bool, error) {
		subj := newSubject(i, TestingSimulationOptions)
		return evalResourceAccessGrantsPrincipal(&subj), nil
	})
}
