package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/trace"
)

// TestEvalIsSameAccount checks same vs x-account checking behavior
func TestEvalIsSameAccount(t *testing.T) {
	type input struct {
		principal entities.Principal
		resource  entities.Resource
	}

	tests := []testrunner.TestCase[input, bool]{
		{
			Input: input{
				principal: entities.Principal{Account: "88888"},
				resource:  entities.Resource{Account: "88888"},
			},
			Want: true,
		},
		{
			Input: input{
				principal: entities.Principal{Account: "88888"},
				resource:  entities.Resource{Account: "12345"},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalIsSameAccount(&i.principal, &i.resource), nil
	})
}

// TestOverallAccess_XAccount checks both principal-side and resource-side logic where the
// resource + principal reside within the same account
func TestOverallAccess_XAccount(t *testing.T) {
	tests := []testrunner.TestCase[AuthContext, bool]{
		{
			Name: "x_account_implicit_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:              "arn:aws:iam::88888:role/myrole",
					Account:          "88888",
					InlinePolicies:   nil,
					AttachedPolicies: nil,
				},
				Resource: &entities.Resource{
					Arn:     "arn:aws:s3:::mybucket",
					Account: "11111",
					Policy:  policy.Policy{},
				},
			},
			Want: false,
		},
		{
			Name: "x_account_principal_only_allow",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:     "arn:aws:iam::88888:role/myrole",
					Account: "88888",
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
					Arn:     "arn:aws:s3:::mybucket",
					Account: "11111",
				},
			},
			Want: false,
		},
		{
			Name: "x_account_resource_only_allow",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:     "arn:aws:iam::88888:role/myrole",
					Account: "88888",
				},
				Resource: &entities.Resource{
					Arn:     "arn:aws:s3:::mybucket",
					Account: "11111",
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
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:     "arn:aws:iam::88888:role/myrole",
					Account: "88888",
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
					Arn:     "arn:aws:s3:::mybucket",
					Account: "11111",
				},
			},
			Want: false,
		},
		{
			Name: "x_account_resource_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:     "arn:aws:iam::88888:role/myrole",
					Account: "88888",
				},
				Resource: &entities.Resource{
					Arn:     "arn:aws:s3:::mybucket",
					Account: "11111",
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
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:     "arn:aws:iam::88888:role/myrole",
					Account: "88888",
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
					Arn:     "arn:aws:s3:::mybucket",
					Account: "11111",
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
			Name: "x_account_error_nonexistent_principal_condition",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:     "arn:aws:iam::88888:role/myrole",
					Account: "88888",
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
					Arn:     "arn:aws:s3:::mybucket",
					Account: "11111",
				},
			},
			ShouldErr: true,
		},
		{
			Name: "x_account_error_nonexistent_resource_condition",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:     "arn:aws:iam::88888:role/myrole",
					Account: "88888",
				},
				Resource: &entities.Resource{
					Arn:     "arn:aws:s3:::mybucket",
					Account: "11111",
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

	testrunner.RunTestSuite(t, tests, func(ac AuthContext) (bool, error) {
		if ac.Principal.Account == ac.Resource.Account {
			t.Fatalf("supposed to be testing x-account, but saw same account for: %+v", ac)
		}

		opts := Options{FailOnUnknownCondition: true}
		res, err := evalOverallAccess(&opts, ac)
		if err != nil {
			return false, err
		}

		return res.IsAllowed, nil
	})
}

// TestOverallAccess_SameAccount checks both principal-side and resource-side logic where the
// resource + principal reside within the same account
func TestOverallAccess_SameAccount(t *testing.T) {
	tests := []testrunner.TestCase[AuthContext, bool]{
		{
			Name: "same_account_implicit_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:              "arn:aws:iam::88888:role/myrole",
					Account:          "88888",
					InlinePolicies:   nil,
					AttachedPolicies: nil,
				},
				Resource: &entities.Resource{
					Arn:     "arn:aws:s3:::mybucket",
					Account: "88888",
					Policy:  policy.Policy{},
				},
			},
			Want: false,
		},
		{
			Name: "same_account_simple_allow",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:     "arn:aws:iam::88888:role/myrole",
					Account: "88888",
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
					Arn:     "arn:aws:s3:::mybucket",
					Account: "88888",
				},
			},
			Want: true,
		},
		{
			Name: "same_account_simple_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:     "arn:aws:iam::88888:role/myrole",
					Account: "88888",
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
					Arn:     "arn:aws:s3:::mybucket",
					Account: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "allow_and_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:     "arn:aws:iam::88888:role/myrole",
					Account: "88888",
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
					Arn:     "arn:aws:s3:::mybucket",
					Account: "88888",
				},
			},
			Want: false,
		},
		{
			Name: "same_account_error_nonexistent_condition",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:     "arn:aws:iam::88888:role/myrole",
					Account: "88888",
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
					Arn:     "arn:aws:s3:::mybucket",
					Account: "88888",
				},
			},
			ShouldErr: true,
		},
		// FIXME(nsiow) uncomment this test when ready for same-account edge case handling
		// {
		// 	Name: "same_account_resource_access",
		// 	Input: AuthContext{
		// 		Action: "s3:listbucket",
		// 		Principal: &entities.Principal{
		// 			Arn: "arn:aws:iam::88888:role/myrole",
		//			Account: "88888",
		// 		},
		// 		Resource: &entities.Resource{
		// 			Arn: "arn:aws:s3:::mybucket",
		//			Account: "88888",
		// 			Policy: policy.Policy{
		// 				Statement: []policy.Statement{
		// 					{
		// 						Effect:   policy.EFFECT_ALLOW,
		// 						Action:   []string{"s3:listbucket"},
		// 						Resource: []string{"arn:aws:s3:::mybucket"},
		// 						Principal: policy.Principal{
		// 							AWS: []string{"arn:aws:iam::88888:role/myrole"},
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 	},
		// 	Want: []policy.Effect{policy.EFFECT_ALLOW},
		// },
	}

	testrunner.RunTestSuite(t, tests, func(ac AuthContext) (bool, error) {
		if ac.Principal.Account != ac.Resource.Account {
			t.Fatalf("supposed to be testing same account, but saw x-account for: %+v", ac)
		}

		opts := Options{FailOnUnknownCondition: true}
		res, err := evalOverallAccess(&opts, ac)
		if err != nil {
			return false, err
		}

		return res.IsAllowed, nil
	})
}

// TestPrincipalAccess checks identity-policy evaluation logic for statements
func TestPrincipalAccess(t *testing.T) {
	tests := []testrunner.TestCase[AuthContext, []policy.Effect]{
		{
			Name: "implicit_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:              "arn:aws:iam::88888:role/myrole",
					InlinePolicies:   nil,
					AttachedPolicies: nil,
				},
				Resource: &entities.Resource{
					Arn: "arn:aws:s3:::mybucket",
				},
			},
			Want: []policy.Effect(nil),
		},
		{
			Name: "simple_inline_policy",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
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
					Arn: "arn:aws:s3:::mybucket",
				},
			},
			Want: []policy.Effect{policy.EFFECT_ALLOW},
		},
		{
			Name: "simple_attached_policy",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
					AttachedPolicies: []policy.Policy{
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
					Arn: "arn:aws:s3:::mybucket",
				},
			},
			Want: []policy.Effect{policy.EFFECT_ALLOW},
		},
		{
			Name: "simple_inline_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
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
					Arn: "arn:aws:s3:::mybucket",
				},
			},
			Want: []policy.Effect{policy.EFFECT_DENY},
		},
		{
			Name: "simple_attached_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
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
					Arn: "arn:aws:s3:::mybucket",
				},
			},
			Want: []policy.Effect{policy.EFFECT_DENY},
		},
		{
			Name: "allow_and_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
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
					Arn: "arn:aws:s3:::mybucket",
				},
			},
			Want: []policy.Effect{policy.EFFECT_ALLOW, policy.EFFECT_DENY},
		},
		{
			Name: "error_nonexistent_condition",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
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
					Arn: "arn:aws:s3:::mybucket",
				},
			},
			ShouldErr: true,
		},
	}

	testrunner.RunTestSuite(t, tests, func(ac AuthContext) ([]policy.Effect, error) {
		opts := Options{FailOnUnknownCondition: true}
		res, err := evalPrincipalAccess(trace.New(), &opts, ac)
		if err != nil {
			return nil, err
		}

		return res.Effects(), nil
	})
}

// TestResourceAccess checks resource-policy evaluation logic for statements
func TestResourceAccess(t *testing.T) {
	tests := []testrunner.TestCase[AuthContext, []policy.Effect]{
		{
			Name: "implicit_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: &entities.Resource{
					Arn:    "arn:aws:s3:::mybucket",
					Policy: policy.Policy{},
				},
			},
			Want: []policy.Effect(nil),
		},
		{
			Name: "simple_match",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: &entities.Resource{
					Arn: "arn:aws:s3:::mybucket",
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
			Want: []policy.Effect{policy.EFFECT_ALLOW},
		},
		{
			Name: "explicit_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: &entities.Resource{
					Arn: "arn:aws:s3:::mybucket",
					Policy: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_DENY,
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
			Want: []policy.Effect{policy.EFFECT_DENY},
		},
		{
			Name: "allow_and_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: &entities.Resource{
					Arn: "arn:aws:s3:::mybucket",
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
							{
								Effect:   policy.EFFECT_DENY,
								Action:   []string{"s3:listbucket"},
								Resource: []string{"arn:aws:s3:::mybucket"},
								Principal: policy.Principal{
									AWS: []string{"*"},
								},
							},
						},
					},
				},
			},
			Want: []policy.Effect{policy.EFFECT_ALLOW, policy.EFFECT_DENY},
		},
		{
			Name: "error_nonexistent_condition",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: &entities.Resource{
					Arn: "arn:aws:s3:::mybucket",
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

	testrunner.RunTestSuite(t, tests, func(ac AuthContext) ([]policy.Effect, error) {
		opts := Options{FailOnUnknownCondition: true}
		res, err := evalResourceAccess(trace.New(), &opts, ac)
		if err != nil {
			return nil, err
		}

		return res.Effects(), nil
	})
}

// TestStatementMatchesAction checks action-matching logic for statements
func TestStatementMatchesAction(t *testing.T) {
	type input struct {
		ac   AuthContext
		stmt policy.Statement
	}

	tests := []testrunner.TestCase[input, bool]{
		// Action
		{
			Name: "simple_wildcard",
			Input: input{
				ac:   AuthContext{Action: "s3:getobject"},
				stmt: policy.Statement{Action: []string{"*"}},
			},
			Want: true,
		},
		{
			Name: "simple_direct_match",
			Input: input{
				ac:   AuthContext{Action: "s2:getobject"},
				stmt: policy.Statement{Action: []string{"s2:getobject"}},
			},
			Want: true,
		},
		{
			Name: "other_action",
			Input: input{
				ac:   AuthContext{Action: "s3:putobject"},
				stmt: policy.Statement{Action: []string{"s3:getobject"}},
			},
			Want: false,
		},
		{
			Name: "two_actions",
			Input: input{
				ac:   AuthContext{Action: "s3:getobject"},
				stmt: policy.Statement{Action: []string{"s3:putobject", "s3:getobject"}},
			},
			Want: true,
		},
		{
			Name: "diff_casing",
			Input: input{
				ac:   AuthContext{Action: "s3:gEtObJeCt"},
				stmt: policy.Statement{Action: []string{"s3:putobject", "s3:getobject"}},
			},
			Want: true,
		},

		// NotAction
		{
			Name: "notaction_simple_wildcard",
			Input: input{
				ac:   AuthContext{Action: "s3:getobject"},
				stmt: policy.Statement{NotAction: []string{"*"}},
			},
			Want: false,
		},
		{
			Name: "notaction_simple_direct_match",
			Input: input{
				ac:   AuthContext{Action: "s3:getobject"},
				stmt: policy.Statement{NotAction: []string{"s3:getobject"}},
			},
			Want: false,
		},
		{
			Name: "notaction_other_action",
			Input: input{
				ac:   AuthContext{Action: "sqs:sendmessage"},
				stmt: policy.Statement{NotAction: []string{"s3:getobject"}},
			},
			Want: true,
		},
		{
			Name: "notaction_two_actions",
			Input: input{
				ac:   AuthContext{Action: "s3:getobject"},
				stmt: policy.Statement{NotAction: []string{"s3:putobject", "s3:getobject"}},
			},
			Want: false,
		},
		{
			Name: "notaction_diff_casing",
			Input: input{
				ac:   AuthContext{Action: "s3:gEtObJeCt"},
				stmt: policy.Statement{NotAction: []string{"s3:putobject", "s3:getobject"}},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesAction(trace.New(), &Options{}, i.ac, &i.stmt)
	})
}

// TestStatementMatchesPrincipal checks principal-matching logic for statements
func TestStatementMatchesPrincipal(t *testing.T) {
	type input struct {
		ac   AuthContext
		stmt policy.Statement
	}

	tests := []testrunner.TestCase[input, bool]{
		// Principal
		{
			Name: "simple_wildcard",
			Input: input{
				ac:   AuthContext{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"*"}}},
			},
			Want: true,
		},
		{
			Name: "simple_direct_match",
			Input: input{
				ac:   AuthContext{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerole"}}},
			},
			Want: true,
		},
		{
			Name: "other_principal",
			Input: input{
				ac:   AuthContext{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerandomrole"}}},
			},
			Want: false,
		},
		{
			Name: "two_principals",
			Input: input{
				ac: AuthContext{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/secondrole"}},
				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{
					"arn:aws:iam::88888:role/firstrole",
					"arn:aws:iam::88888:role/secondrole"}}}},
			Want: true,
		},
		{
			Name: "other_service",
			Input: input{
				ac:   AuthContext{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{Principal: policy.Principal{Federated: []string{"*"}}},
			},
			Want: false,
		},

		// NotPrincipal
		{
			Name: "notprincipal_simple_wildcard",
			Input: input{
				ac:   AuthContext{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"*"}}},
			},
			Want: false,
		},
		{
			Name: "notprincipal_simple_direct_match",
			Input: input{
				ac:   AuthContext{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerole"}}},
			},
			Want: false,
		},
		{
			Name: "notprincipal_other_principal",
			Input: input{
				ac:   AuthContext{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerandomrole"}}},
			},
			Want: true,
		},
		{
			Name: "notprincipal_two_principals",
			Input: input{
				ac: AuthContext{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/secondrole"}},
				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{
					"arn:aws:iam::88888:role/firstrole",
					"arn:aws:iam::88888:role/secondrole"}}}},
			Want: false,
		},
		{
			Name: "notprincipal_other_service",
			Input: input{
				ac:   AuthContext{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{NotPrincipal: policy.Principal{Federated: []string{"*"}}},
			},
			Want: true,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesPrincipal(trace.New(), &Options{}, i.ac, &i.stmt)
	})
}

// TestStatementMatchesResource checks resource-matching logic for statements
func TestStatementMatchesResource(t *testing.T) {
	type input struct {
		ac   AuthContext
		stmt policy.Statement
	}

	tests := []testrunner.TestCase[input, bool]{
		// Resource
		{
			Name: "simple_wildcard",
			Input: input{
				ac:   AuthContext{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
				stmt: policy.Statement{Resource: []string{"*"}},
			},
			Want: true,
		},
		{
			Name: "simple_direct_match",
			Input: input{
				ac:   AuthContext{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
				stmt: policy.Statement{Resource: []string{"arn:aws:s3:::somebucket"}},
			},
			Want: true,
		},
		{
			Name: "other_resource",
			Input: input{
				ac:   AuthContext{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
				stmt: policy.Statement{Resource: []string{"arn:aws:s3:::adifferentbucket"}},
			},
			Want: false,
		},
		{
			Name: "two_resources",
			Input: input{
				ac: AuthContext{Resource: &entities.Resource{Arn: "arn:aws:s3:::secondbucket"}},
				stmt: policy.Statement{Resource: []string{
					"arn:aws:s3:::firstbucket",
					"arn:aws:s3:::secondbucket"}},
			},
			Want: true,
		},

		// NotResource
		{
			Name: "notresource_simple_wildcard",
			Input: input{
				ac:   AuthContext{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
				stmt: policy.Statement{NotResource: []string{"*"}},
			},
			Want: false,
		},
		{
			Name: "notresource_simple_direct_match",
			Input: input{
				ac:   AuthContext{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
				stmt: policy.Statement{NotResource: []string{"arn:aws:s3:::somebucket"}},
			},
			Want: false,
		},
		{
			Name: "notresource_other_resource",
			Input: input{
				ac:   AuthContext{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
				stmt: policy.Statement{NotResource: []string{"arn:aws:s3:::adifferentbucket"}},
			},
			Want: true,
		},
		{
			Name: "notresource_two_resources",
			Input: input{
				ac: AuthContext{Resource: &entities.Resource{Arn: "arn:aws:s3:::secondbucket"}},
				stmt: policy.Statement{NotResource: []string{
					"arn:aws:s3:::firstbucket",
					"arn:aws:s3:::secondbucket"}},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesResource(trace.New(), &Options{}, i.ac, &i.stmt)
	})
}

// TestPermissionsBoundary tests functionality of permissions boundary evaluations
func TestPermissionsBoundary(t *testing.T) {
	tests := []testrunner.TestCase[AuthContext, []policy.Effect]{
		{
			Name: "allow_all",
			Input: AuthContext{
				Principal: &entities.Principal{
					PermissionsBoundary: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_ALLOW,
								Action:   []string{"*"},
								Resource: []string{"*"},
							},
						},
					},
				},
				Resource: &entities.Resource{Arn: "arn:aws:s3:::mybucket"},
			},
			Want: []policy.Effect{
				policy.EFFECT_ALLOW,
			},
		},
	}

	testrunner.RunTestSuite(t, tests, func(ac AuthContext) ([]policy.Effect, error) {
		res, err := evalPermissionsBoundary(trace.New(), &Options{}, ac)
		if err != nil {
			return nil, err
		}

		return res.Effects(), nil
	})
}
