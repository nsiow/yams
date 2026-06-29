package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
	"github.com/nsiow/yams/pkg/policy"
)

func TestEvalIsSameAccount(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, bool]{
		{
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					AccountId: "88888",
				},
				Resource: &entities.FrozenResource{
					AccountId: "88888",
				},
			},
			Want: true,
		},
		{
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					AccountId: "12345",
				},
				Resource: &entities.FrozenResource{
					AccountId: "88888",
				},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i AuthContext) (bool, error) {
		subj := newSubject(i, TestingSimulationOptions)
		return evalIsSameAccount(&subj), nil
	})
}

func TestOverallAccess_XAccount(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, bool]{
		{
			Name: "x_account_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{

					Arn:              "arn:aws:iam::88888:role/myrole",
					AccountId:        "88888",
					InlinePolicies:   nil,
					AttachedPolicies: nil,
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "11111",
					Policy:    policy.Policy{},
				},
			},
			Want: false,
		},
		{
			Name: "x_account_principal_only_allow",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{

					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Version: "2012-10-17",
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "11111",
				},
			},
			Want: false,
		},
		{
			Name: "x_account_resource_only_allow",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{

					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "11111",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_ALLOW,
								Action:   []string{"s3:listbucket"},
								Resource: []string{"arn:aws:s3:::mybucket"},
								Principal: policy.Principal{
									AWS: []string{"arn:aws:iam::88888:role/myrole"},
								},
							},
						},
					},
				},
			},
			Want: false,
		},
		{

			Name: "x_account_principal_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{

					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_DENY,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "11111",
				},
			},
			Want: false,
		},
		{
			Name: "x_account_resource_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{

					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "11111",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_DENY,
								Action:   []string{"s3:listbucket"},
								Resource: []string{"arn:aws:s3:::mybucket"},
							},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "x_account_allow_and_allow",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{

					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "11111",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_ALLOW,
								Action:   []string{"s3:listbucket"},
								Resource: []string{"arn:aws:s3:::mybucket"},
								Principal: policy.Principal{
									AWS: []string{"arn:aws:iam::88888:role/myrole"},
								},
							},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "nonexistent_principal_condition",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{

					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
									Principal: policy.Principal{
										AWS: []string{"arn:aws:iam::88888:role/myrole"},
									},
									Condition: map[string]map[string]policy.Value{
										"StringEqualsThisDoesNotExist": {
											"foo": []string{"bar"},
										},
									},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "11111",
				},
			},
			Want: false,
		},
		{
			Name: "nonexistent_resource_condition",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{

					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "11111",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_ALLOW,
								Action:   []string{"s3:listbucket"},
								Resource: []string{"arn:aws:s3:::mybucket"},
								Principal: policy.Principal{
									AWS: []string{"arn:aws:iam::88888:role/myrole"},
								},
								Condition: map[string]map[string]policy.Value{
									"StringEqualsThisDoesNotExist": {
										"foo": []string{"bar"},
									},
								},
							},
						},
					},
				},
			},
			Want: false,
		},
		{
			// account-root delegation should be recognized in non-aws partitions
			// (aws-cn, aws-us-gov), not just the default aws partition.
			Name: "x_account_allow_and_allow_aws_cn_partition",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws-cn:iam::99999:role/myrole",
					AccountId: "99999",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws-cn:s3:::mybucket"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws-cn:s3:::mybucket",
					AccountId: "11111",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect: policy.EFFECT_ALLOW,
								Principal: policy.Principal{
									AWS: []string{"arn:aws-cn:iam::99999:root"},
								},
								Action:   []string{"s3:listbucket"},
								Resource: []string{"arn:aws-cn:s3:::mybucket"},
							},
						},
					},
				},
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) (bool, error) {
		if ac.Principal.AccountId == ac.Resource.AccountId {
			t.Fatalf("supposed to be testing x-account, but saw same account for: %+v", ac)
		}

		subj := newSubject(ac, TestingSimulationOptions)
		res := evalOverallAccess(&subj)
		return res.IsAllowed, nil
	})
}

func TestOverallAccess_SameAccount(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, bool]{
		{
			Name: "same_account_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:              "arn:aws:iam::88888:role/myrole",
					AccountId:        "88888",
					InlinePolicies:   nil,
					AttachedPolicies: nil,
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
					Policy:    policy.Policy{},
				},
			},
			Want: false,
		},
		{
			Name: "same_account_simple_allow",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: true,
		},
		{
			Name: "same_account_strict_allow",
			Input: AuthContext{
				Action: sar.MustLookupString("sts:assumerole"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"sts:AssumeRole"},
									Resource: []string{"*"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:iam::88888:role/yourrole",
					Type:      "AWS::IAM::Role",
					AccountId: "88888",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect: policy.EFFECT_ALLOW,
								Principal: policy.Principal{
									AWS: policy.Value{
										"88888",
									},
								},
							},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "same_account_strict_deny_no_resource_policy",
			Input: AuthContext{
				Action: sar.MustLookupString("sts:assumerole"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"sts:AssumeRole"},
									Resource: []string{"*"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:iam::88888:role/yourrole",
					Type:      "AWS::IAM::Role",
					AccountId: "88888",
					// No trust policy - resource doesn't allow
				},
			},
			Want: false,
		},
		{
			Name: "same_account_sts_role_action_requires_resource_policy",
			Input: AuthContext{
				Action: sar.MustLookupString("sts:assumerolewithsaml"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"sts:AssumeRoleWithSAML"},
									Resource: []string{"*"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:iam::88888:role/yourrole",
					Type:      "AWS::IAM::Role",
					AccountId: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "same_account_simple_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Version: "2012-10-17",
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_DENY,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "allow_and_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
					AttachedPolicies: []entities.ManagedPolicy{
						{
							Policy: policy.Policy{
								Statement: []policy.Statement{
									{
										Effect:   policy.EFFECT_DENY,
										Action:   []string{"s3:listbucket"},
										Resource: []string{"arn:aws:s3:::mybucket"},
									},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "same_account_nonexistent_condition",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
									Principal: policy.Principal{
										AWS: []string{"arn:aws:iam::88888:role/myrole"},
									},
									Condition: map[string]map[string]policy.Value{
										"StringEqualsThisDoesNotExist": {
											"foo": []string{"bar"},
										},
									},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "same_account_resource_access",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_ALLOW,
								Action:   []string{"s3:listbucket"},
								Resource: []string{"arn:aws:s3:::mybucket"},
								Principal: policy.Principal{
									AWS: []string{"arn:aws:iam::88888:role/myrole"},
								},
							},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "same_account_resource_access_role_boundary_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Type:      awsconfig.CONST_TYPE_AWS_IAM_ROLE,
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					PermissionBoundary: entities.ManagedPolicy{
						Arn: "arn:aws:iam::88888:policy/boundary",
						Policy: policy.Policy{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"ec2:*"},
									Resource: []string{"*"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_ALLOW,
								Action:   []string{"s3:listbucket"},
								Resource: []string{"arn:aws:s3:::mybucket"},
								Principal: policy.Principal{
									AWS: []string{"arn:aws:iam::88888:role/myrole"},
								},
							},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "same_account_resource_access_user_boundary_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Type:      awsconfig.CONST_TYPE_AWS_IAM_USER,
					Arn:       "arn:aws:iam::88888:user/myuser",
					AccountId: "88888",
					PermissionBoundary: entities.ManagedPolicy{
						Arn: "arn:aws:iam::88888:policy/boundary",
						Policy: policy.Policy{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"ec2:*"},
									Resource: []string{"*"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_ALLOW,
								Action:   []string{"s3:listbucket"},
								Resource: []string{"arn:aws:s3:::mybucket"},
								Principal: policy.Principal{
									AWS: []string{"arn:aws:iam::88888:user/myuser"},
								},
							},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "permissions_boundary_allow",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
					PermissionBoundary: entities.ManagedPolicy{
						Policy: policy.Policy{
							Statement: []policy.Statement{
								{
									Effect:    policy.EFFECT_ALLOW,
									NotAction: []string{"iam:*"},
									Resource:  []string{"*"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: true,
		},
		{
			Name: "permissions_boundary_explicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
					PermissionBoundary: entities.ManagedPolicy{
						Policy: policy.Policy{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_DENY,
									Action:   []string{"*"},
									Resource: []string{"*"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "permissions_boundary_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
					PermissionBoundary: entities.ManagedPolicy{
						Policy: policy.Policy{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"ec2:*"},
									Resource: []string{"*"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "scp_allow",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect:   policy.EFFECT_ALLOW,
													Action:   []string{"*"},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: true,
		},
		{
			Name: "scp_explicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect:   policy.EFFECT_ALLOW,
													Action:   []string{"*"},
													Resource: []string{"*"},
												},
											},
										},
									},
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect:   policy.EFFECT_DENY,
													Action:   []string{"*"},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "scp_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect:   policy.EFFECT_ALLOW,
													Action:   []string{"ec2:*"},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "scp_layer_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect: policy.EFFECT_ALLOW,
													Action: []string{"*"},
													Principal: policy.Principal{
														AWS: []string{"*"},
													},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
							{
								SCPs: []entities.ManagedPolicy{},
							},
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect: policy.EFFECT_ALLOW,
													Action: []string{"*"},
													Principal: policy.Principal{
														AWS: []string{"*"},
													},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "rcp_allow",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Type:      "AWS::S3::Bucket",
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								RCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect: policy.EFFECT_ALLOW,
													Action: []string{"*"},
													Principal: policy.Principal{
														AWS: []string{"*"},
													},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "rcp_explicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Type:      "AWS::S3::Bucket",
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								RCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect:   policy.EFFECT_ALLOW,
													Action:   []string{"*"},
													Resource: []string{"*"},
												},
											},
										},
									},
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect:   policy.EFFECT_DENY,
													Action:   []string{"*"},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "rcp_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Type:      "AWS::S3::Bucket",
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								RCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect:   policy.EFFECT_ALLOW,
													Action:   []string{"ec2:*"},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "rcp_layer_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Type:      "AWS::S3::Bucket",
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								RCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect:   policy.EFFECT_ALLOW,
													Action:   []string{"*"},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
							{
								RCPs: []entities.ManagedPolicy{},
							},
							{
								RCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect:   policy.EFFECT_ALLOW,
													Action:   []string{"*"},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Want: false,
		},
		{
			// resource-grants-principal short-circuit must verify the policy's Resource
			// block actually targets the requested resource. A bucketA policy that grants
			// access to bucketB should not let the principal reach bucketA without an
			// identity policy.
			Name: "resource_grants_principal_wrong_resource",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::bucketA",
					AccountId: "88888",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect: policy.EFFECT_ALLOW,
								Principal: policy.Principal{
									AWS: []string{"arn:aws:iam::88888:role/myrole"},
								},
								Action:   []string{"s3:listbucket"},
								Resource: []string{"arn:aws:s3:::bucketB"},
							},
						},
					},
				},
			},
			Want: false,
		},
		{
			// SCPs may use NotPrincipal to exempt specific principals from a deny.
			// https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_scps_syntax.html
			Name: "scp_not_principal_exempts",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/exempt-role",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect:   policy.EFFECT_ALLOW,
													Action:   []string{"*"},
													Resource: []string{"*"},
												},
											},
										},
									},
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect: policy.EFFECT_DENY,
													NotPrincipal: policy.Principal{
														AWS: []string{
															"arn:aws:iam::88888:role/exempt-role",
														},
													},
													Action:   []string{"s3:*"},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: true,
		},
		{
			// SCP early-return must inspect every node, not just OrgNodes[0]. A deny SCP
			// at the account level should still apply when the root has no SCPs of its own.
			Name: "scp_deny_at_account_when_root_empty",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{Name: "Root", Type: "ROOT"},
							{
								Name: "Account",
								Type: "ACCOUNT",
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect:   policy.EFFECT_DENY,
													Action:   []string{"s3:*"},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: false,
		},
		{
			// Same shape of bug as above, but for RCPs on the resource side.
			Name: "rcp_deny_at_account_when_root_empty",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Type:      "AWS::S3::Bucket",
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{Name: "Root", Type: "ROOT"},
							{
								Name: "Account",
								Type: "ACCOUNT",
								RCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect:   policy.EFFECT_DENY,
													Action:   []string{"s3:*"},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Want: false,
		},
		{
			// IAM variables in the Resource element should be substituted before matching.
			// https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_variables.html
			Name: "resource_arn_variable_substituted",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:getobject"),
				Principal: &entities.FrozenPrincipal{
					Arn:       "arn:aws:iam::88888:user/alice",
					AccountId: "88888",
					Type:      "AWS::IAM::User",
					InlinePolicies: []policy.Policy{
						{
							Version: "2012-10-17",
							Statement: []policy.Statement{
								{
									Effect: policy.EFFECT_ALLOW,
									Action: []string{"s3:GetObject"},
									Resource: []string{
										"arn:aws:s3:::mybucket/${aws:username}/*",
									},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn:         "arn:aws:s3:::mybucket/alice/myfile",
					AccountId:   "88888",
					ArnSegments: entities.SplitArn("arn:aws:s3:::mybucket/alice/myfile"),
				},
				Properties: NewBagFromMap(map[string]string{
					"aws:username": "alice",
				}),
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) (bool, error) {
		if ac.Principal.AccountId != ac.Resource.AccountId {
			t.Fatalf("supposed to be testing same account, but saw x-account for: %+v", ac)
		}

		subj := newSubject(ac, TestingSimulationOptions)
		res := evalOverallAccess(&subj)
		return res.IsAllowed, nil
	})
}
