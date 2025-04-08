package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestEvalIsSameAccount(t *testing.T) {
	type input struct {
		principal *entities.Principal
		resource  *entities.Resource
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Input: input{
				principal: &entities.Principal{AccountId: "88888"},
				resource:  &entities.Resource{AccountId: "88888"},
			},
			Want: true,
		},
		{
			Input: input{
				principal: &entities.Principal{AccountId: "88888"},
				resource:  &entities.Resource{AccountId: "12345"},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalIsSameAccount(i.principal, i.resource), nil
	})
}

func TestOverallAccess_XAccount(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, bool]{
		{
			Name: "x_account_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
					Arn:              "arn:aws:iam::88888:role/myrole",
					AccountId:        "88888",
					InlinePolicies:   nil,
					AttachedPolicies: nil,
				},
				Resource: &entities.Resource{
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
				Principal: &entities.Principal{
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
				Resource: &entities.Resource{
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
				Principal: &entities.Principal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
				},
				Resource: &entities.Resource{
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
				Principal: &entities.Principal{
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
				Resource: &entities.Resource{
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
				Principal: &entities.Principal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
				},
				Resource: &entities.Resource{
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
				Principal: &entities.Principal{
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
				Resource: &entities.Resource{
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
			Name: "error_nonexistent_principal_condition",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
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
				Resource: &entities.Resource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "11111",
				},
			},
			ShouldErr: true,
		},
		{
			Name: "error_nonexistent_resource_condition",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
				},
				Resource: &entities.Resource{
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
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) (bool, error) {
		if ac.Principal.AccountId == ac.Resource.AccountId {
			t.Fatalf("supposed to be testing x-account, but saw same account for: %+v", ac)
		}

		subj := newSubject(&ac, TestingSimulationOptions)
		res, err := evalOverallAccess(subj)
		if err != nil {
			return false, err
		}

		return res.IsAllowed, nil
	})
}

func TestOverallAccess_SameAccount(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, bool]{
		{
			Name: "same_account_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
					Arn:              "arn:aws:iam::88888:role/myrole",
					AccountId:        "88888",
					InlinePolicies:   nil,
					AttachedPolicies: nil,
				},
				Resource: &entities.Resource{
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
				Principal: &entities.Principal{
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
				Resource: &entities.Resource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: true,
		},
		{
			Name: "same_account_simple_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
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
				Resource: &entities.Resource{
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
				Principal: &entities.Principal{
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
					AttachedPolicies: []policy.Policy{
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
				Resource: &entities.Resource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "same_account_error_nonexistent_condition",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
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
				Resource: &entities.Resource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			ShouldErr: true,
		},
		{
			Name: "same_account_resource_access",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
				},
				Resource: &entities.Resource{
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
			Name: "error_bad_permissions_boundary",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					PermissionsBoundary: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_ALLOW,
								Action:   []string{"s3:listbucket"},
								Resource: []string{"arn:aws:s3:::mybucket"},
								Condition: map[string]map[string]policy.Value{
									"StringEqualsThisDoesNotExist": {
										"foo": []string{"bar"},
									},
								},
							},
						},
					},
				},
				Resource: &entities.Resource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			ShouldErr: true,
		},
		{
			Name: "permissions_boundary_allow",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
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
					PermissionsBoundary: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:    policy.EFFECT_ALLOW,
								NotAction: []string{"iam:*"},
								Resource:  []string{"*"},
							},
						},
					},
				},
				Resource: &entities.Resource{
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
				Principal: &entities.Principal{
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
					PermissionsBoundary: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_DENY,
								Action:   []string{"*"},
								Resource: []string{"*"},
							},
						},
					},
				},
				Resource: &entities.Resource{
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
				Principal: &entities.Principal{
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
					PermissionsBoundary: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_ALLOW,
								Action:   []string{"ec2:*"},
								Resource: []string{"*"},
							},
						},
					},
				},
				Resource: &entities.Resource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "error_bad_scp",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					Account: entities.Account{
						SCPs: [][]policy.Policy{
							{
								{
									Statement: []policy.Statement{
										{
											Effect:   policy.EFFECT_ALLOW,
											Action:   []string{"*"},
											Resource: []string{"*"},
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
					},
				},
				Resource: &entities.Resource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			ShouldErr: true,
		},
		{
			Name: "scp_allow",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
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
					Account: entities.Account{
						SCPs: [][]policy.Policy{
							{
								{
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
				Resource: &entities.Resource{
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
				Principal: &entities.Principal{
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
					Account: entities.Account{
						SCPs: [][]policy.Policy{
							{
								{
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
								{
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
				Resource: &entities.Resource{
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
				Principal: &entities.Principal{
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
					Account: entities.Account{
						SCPs: [][]policy.Policy{
							{
								{
									Statement: []policy.Statement{
										{
											Effect:    policy.EFFECT_ALLOW,
											NotAction: []string{"s3:*"},
											Resource:  []string{"*"},
										},
									},
								},
							},
						},
					},
				},
				Resource: &entities.Resource{
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
				Principal: &entities.Principal{
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
					Account: entities.Account{
						SCPs: [][]policy.Policy{
							{
								{
									Statement: []policy.Statement{
										{
											Effect:    policy.EFFECT_ALLOW,
											NotAction: []string{"*"},
											Resource:  []string{"*"},
										},
									},
								},
							},
							{}, // <= missing
							{
								{
									Statement: []policy.Statement{
										{
											Effect:    policy.EFFECT_ALLOW,
											NotAction: []string{"*"},
											Resource:  []string{"*"},
										},
									},
								},
							},
						},
					},
				},
				Resource: &entities.Resource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) (bool, error) {
		if ac.Principal.AccountId != ac.Resource.AccountId {
			t.Fatalf("supposed to be testing same account, but saw x-account for: %+v", ac)
		}

		subj := newSubject(&ac, TestingSimulationOptions)
		res, err := evalOverallAccess(subj)
		if err != nil {
			return false, err
		}

		return res.IsAllowed, nil
	})
}
