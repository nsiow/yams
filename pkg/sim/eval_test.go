package sim

import (
	"reflect"
	"testing"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// TestOverallAccess_XAccount checks both principal-side and resource-side logic where the
// resource + principal reside within the same account
func TestOverallAccess_XAccount(t *testing.T) {
	type test struct {
		name  string
		event Event
		want  bool
		opts  Options
		err   bool
	}

	tests := []test{
		{
			name: "x_account_implicit_deny",
			event: Event{
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
			want: false,
		},
		{
			name: "x_account_principal_only_allow",
			event: Event{
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
			want: false,
		},
		{
			name: "x_account_resource_only_allow",
			event: Event{
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
			want: false,
		},
		{

			name: "x_account_principal_deny",
			event: Event{
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
			want: false,
		},
		{
			name: "x_account_resource_deny",
			event: Event{
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
			want: false,
		},
		{
			name: "x_account_allow_and_allow",
			event: Event{
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
			want: true,
		},
		{
			name: "x_account_error_nonexistent_principal_condition",
			event: Event{
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
										"StringEqualsThisDoesNotExist": nil,
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
			opts: Options{FailOnUnknownCondition: true},
			err:  true,
		},
		{
			name: "x_account_error_nonexistent_resource_condition",
			event: Event{
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
									"StringEqualsThisDoesNotExist": nil,
								},
							},
						},
					},
				},
			},
			opts: Options{FailOnUnknownCondition: true},
			err:  true,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		res, err := evalOverallAccess(&tc.opts, &tc.event)

		check := true
		switch {
		case err == nil && tc.err:
			t.Fatalf("expected error, got success for test case '%s': %v", tc.name, err)
		case err != nil && tc.err:
			// expected error; got error
			t.Logf("test saw expected error: %v", err)
			check = false
		case err == nil && !tc.err:
			// no error and not expecting one, continue
		case err != nil && !tc.err:
			t.Fatalf("unable to create policy from test case '%s': %v", tc.name, err)
		}

		if check && !reflect.DeepEqual(res.IsAllowed, tc.want) {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, res.IsAllowed)
		}
	}
}

// TestOverallAccess_SameAccount checks both principal-side and resource-side logic where the
// resource + principal reside within the same account
func TestOverallAccess_SameAccount(t *testing.T) {
	type test struct {
		name  string
		event Event
		want  bool
		opts  Options
		err   bool
	}

	tests := []test{
		{
			name: "same_account_implicit_deny",
			event: Event{
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
			want: false,
		},
		{
			name: "same_account_simple_allow",
			event: Event{
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
			want: true,
		},
		{
			name: "same_account_simple_deny",
			event: Event{
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
			want: false,
		},
		{
			name: "allow_and_deny",
			event: Event{
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
			want: false,
		},
		{
			name: "same_account_error_nonexistent_condition",
			event: Event{
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
										"StringEqualsThisDoesNotExist": nil,
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
			opts: Options{FailOnUnknownCondition: true},
			err:  true,
		},
		// FIXME(nsiow) uncomment this test when ready for same-account edge case handling
		// {
		// 	name: "same_account_resource_access",
		// 	event: Event{
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
		// 	want: []policy.Effect{policy.EFFECT_ALLOW},
		// },
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		res, err := evalOverallAccess(&tc.opts, &tc.event)

		check := true
		switch {
		case err == nil && tc.err:
			t.Fatalf("expected error, got success for test case '%s': %v", tc.name, err)
		case err != nil && tc.err:
			// expected error; got error
			t.Logf("test saw expected error: %v", err)
			check = false
		case err == nil && !tc.err:
			// no error and not expecting one, continue
		case err != nil && !tc.err:
			t.Fatalf("unable to create policy from test case '%s': %v", tc.name, err)
		}

		if check && !reflect.DeepEqual(res.IsAllowed, tc.want) {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, res.IsAllowed)
		}
	}
}

// TestPrincipalAccess checks identity-policy evaluation logic for statements
func TestPrincipalAccess(t *testing.T) {
	type test struct {
		name  string
		event Event
		want  []policy.Effect
		opts  Options
		err   bool
	}

	tests := []test{
		{
			name: "implicit_deny",
			event: Event{
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
			want: []policy.Effect(nil),
		},
		{
			name: "simple_inline_policy",
			event: Event{
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
			want: []policy.Effect{policy.EFFECT_ALLOW},
		},
		{
			name: "simple_attached_policy",
			event: Event{
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
			want: []policy.Effect{policy.EFFECT_ALLOW},
		},
		{
			name: "simple_inline_deny",
			event: Event{
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
			want: []policy.Effect{policy.EFFECT_DENY},
		},
		{
			name: "simple_attached_deny",
			event: Event{
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
			want: []policy.Effect{policy.EFFECT_DENY},
		},
		{
			name: "allow_and_deny",
			event: Event{
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
			want: []policy.Effect{policy.EFFECT_ALLOW, policy.EFFECT_DENY},
		},
		{
			name: "error_nonexistent_condition",
			event: Event{
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
										"StringEqualsThisDoesNotExist": nil,
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
			opts: Options{FailOnUnknownCondition: true},
			err:  true,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		trc := Trace{}
		res, err := evalPrincipalAccess(&tc.opts, &tc.event, &trc)

		check := true
		switch {
		case err == nil && tc.err:
			t.Fatalf("expected error, got success for test case '%s': %v", tc.name, err)
		case err != nil && tc.err:
			// expected error; got error
			t.Logf("test saw expected error: %v", err)
			check = false
		case err == nil && !tc.err:
			// no error and not expecting one, continue
		case err != nil && !tc.err:
			t.Fatalf("unable to create policy from test case '%s': %v", tc.name, err)
		}

		if check && !reflect.DeepEqual(res.Effects(), tc.want) {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, res.Effects())
		}
	}
}

// TestResourceAccess checks resource-policy evaluation logic for statements
func TestResourceAccess(t *testing.T) {
	type test struct {
		name  string
		event Event
		want  []policy.Effect
		opts  Options
		err   bool
	}

	tests := []test{
		{
			name: "implicit_deny",
			event: Event{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: &entities.Resource{
					Arn:    "arn:aws:s3:::mybucket",
					Policy: policy.Policy{},
				},
			},
			want: []policy.Effect(nil),
		},
		{
			name: "simple_match",
			event: Event{
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
			want: []policy.Effect{policy.EFFECT_ALLOW},
		},
		{
			name: "explicit_deny",
			event: Event{
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
			want: []policy.Effect{policy.EFFECT_DENY},
		},
		{
			name: "allow_and_deny",
			event: Event{
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
			want: []policy.Effect{policy.EFFECT_ALLOW, policy.EFFECT_DENY},
		},
		{
			name: "error_nonexistent_condition",
			event: Event{
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
									"StringEqualsThisDoesNotExist": nil,
								},
							},
						},
					},
				},
			},
			opts: Options{FailOnUnknownCondition: true},
			err:  true,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		trc := Trace{}
		res, err := evalResourceAccess(&tc.opts, &tc.event, &trc)

		check := true
		switch {
		case err == nil && tc.err:
			t.Fatalf("expected error, got success for test case '%s': %v", tc.name, err)
		case err != nil && tc.err:
			// expected error; got error
			t.Logf("test saw expected error: %v", err)
			check = false
		case err == nil && !tc.err:
			// no error and not expecting one, continue
		case err != nil && !tc.err:
			t.Fatalf("unable to create policy from test case '%s': %v", tc.name, err)
		}

		if check && !reflect.DeepEqual(res.Effects(), tc.want) {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, res.Effects())
		}
	}
}

// TestStatementMatchesAction checks action-matching logic for statements
func TestStatementMatchesAction(t *testing.T) {
	type test struct {
		name  string
		event Event
		stmt  policy.Statement
		want  bool
	}

	tests := []test{
		// Action
		{
			name:  "simple_wildcard",
			event: Event{Action: "s3:getobject"},
			stmt:  policy.Statement{Action: []string{"*"}},
			want:  true,
		},
		{
			name:  "simple_direct_match",
			event: Event{Action: "s2:getobject"},
			stmt:  policy.Statement{Action: []string{"s2:getobject"}},
			want:  true,
		},
		{
			name:  "other_action",
			event: Event{Action: "s3:putobject"},
			stmt:  policy.Statement{Action: []string{"s3:getobject"}},
			want:  false,
		},
		{
			name:  "two_actions",
			event: Event{Action: "s3:getobject"},
			stmt:  policy.Statement{Action: []string{"s3:putobject", "s3:getobject"}},
			want:  true,
		},
		{
			name:  "diff_casing",
			event: Event{Action: "s3:gEtObJeCt"},
			stmt:  policy.Statement{Action: []string{"s3:putobject", "s3:getobject"}},
			want:  true,
		},

		// NotAction
		{
			name:  "notaction_simple_wildcard",
			event: Event{Action: "s3:getobject"},
			stmt:  policy.Statement{NotAction: []string{"*"}},
			want:  false,
		},
		{
			name:  "notaction_simple_direct_match",
			event: Event{Action: "s3:getobject"},
			stmt:  policy.Statement{NotAction: []string{"s3:getobject"}},
			want:  false,
		},
		{
			name:  "notaction_other_action",
			event: Event{Action: "sqs:sendmessage"},
			stmt:  policy.Statement{NotAction: []string{"s3:getobject"}},
			want:  true,
		},
		{
			name:  "notaction_two_actions",
			event: Event{Action: "s3:getobject"},
			stmt:  policy.Statement{NotAction: []string{"s3:putobject", "s3:getobject"}},
			want:  false,
		},
		{
			name:  "notaction_diff_casing",
			event: Event{Action: "s3:gEtObJeCt"},
			stmt:  policy.Statement{NotAction: []string{"s3:putobject", "s3:getobject"}},
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		opts := Options{}
		trc := Trace{}
		got, err := evalStatementMatchesAction(&opts, &tc.event, &trc, &tc.stmt)
		if err != nil {
			t.Fatalf("observed unexpected error: %v", err)
		}

		if got != tc.want {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, got)
		}
	}
}

// TestStatementMatchesPrincipal checks principal-matching logic for statements
func TestStatementMatchesPrincipal(t *testing.T) {
	type test struct {
		name  string
		event Event
		stmt  policy.Statement
		want  bool
	}

	tests := []test{
		// Principal
		{
			name:  "simple_wildcard",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
			stmt:  policy.Statement{Principal: policy.Principal{AWS: []string{"*"}}},
			want:  true,
		},
		{
			name:  "simple_direct_match",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
			stmt:  policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerole"}}},
			want:  true,
		},
		{
			name:  "other_principal",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
			stmt:  policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerandomrole"}}},
			want:  false,
		},
		{
			name:  "two_principals",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/secondrole"}},
			stmt: policy.Statement{Principal: policy.Principal{AWS: []string{
				"arn:aws:iam::88888:role/firstrole",
				"arn:aws:iam::88888:role/secondrole"}}},
			want: true,
		},
		{
			name:  "other_service",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
			stmt:  policy.Statement{Principal: policy.Principal{Federated: []string{"*"}}},
			want:  false,
		},

		// NotPrincipal
		{
			name:  "notprincipal_simple_wildcard",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
			stmt:  policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"*"}}},
			want:  false,
		},
		{
			name:  "notprincipal_simple_direct_match",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
			stmt:  policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerole"}}},
			want:  false,
		},
		{
			name:  "notprincipal_other_principal",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
			stmt:  policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerandomrole"}}},
			want:  true,
		},
		{
			name:  "notprincipal_two_principals",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/secondrole"}},
			stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{
				"arn:aws:iam::88888:role/firstrole",
				"arn:aws:iam::88888:role/secondrole"}}},
			want: false,
		},
		{
			name:  "notprincipal_other_service",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
			stmt:  policy.Statement{NotPrincipal: policy.Principal{Federated: []string{"*"}}},
			want:  true,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		opts := Options{}
		trc := Trace{}
		got, err := evalStatementMatchesPrincipal(&opts, &tc.event, &trc, &tc.stmt)
		if err != nil {
			t.Fatalf("observed unexpected error: %v", err)
		}

		if got != tc.want {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, got)
		}
	}
}

// TestStatementMatchesResource checks resource-matching logic for statements
func TestStatementMatchesResource(t *testing.T) {
	type test struct {
		name  string
		event Event
		stmt  policy.Statement
		want  bool
	}

	tests := []test{
		// Resource
		{
			name:  "simple_wildcard",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
			stmt:  policy.Statement{Resource: []string{"*"}},
			want:  true,
		},
		{
			name:  "simple_direct_match",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
			stmt:  policy.Statement{Resource: []string{"arn:aws:s3:::somebucket"}},
			want:  true,
		},
		{
			name:  "other_resource",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
			stmt:  policy.Statement{Resource: []string{"arn:aws:s3:::adifferentbucket"}},
			want:  false,
		},
		{
			name:  "two_resources",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::secondbucket"}},
			stmt: policy.Statement{Resource: []string{
				"arn:aws:s3:::firstbucket",
				"arn:aws:s3:::secondbucket"}},
			want: true,
		},

		// NotResource
		{
			name:  "notresource_simple_wildcard",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
			stmt:  policy.Statement{NotResource: []string{"*"}},
			want:  false,
		},
		{
			name:  "notresource_simple_direct_match",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
			stmt:  policy.Statement{NotResource: []string{"arn:aws:s3:::somebucket"}},
			want:  false,
		},
		{
			name:  "notresource_other_resource",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
			stmt:  policy.Statement{NotResource: []string{"arn:aws:s3:::adifferentbucket"}},
			want:  true,
		},
		{
			name:  "notresource_two_resources",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::secondbucket"}},
			stmt: policy.Statement{NotResource: []string{
				"arn:aws:s3:::firstbucket",
				"arn:aws:s3:::secondbucket"}},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		opts := Options{}
		trc := Trace{}
		got, err := evalStatementMatchesResource(&opts, &tc.event, &trc, &tc.stmt)
		if err != nil {
			t.Fatalf("observed unexpected error: %v", err)
		}

		if got != tc.want {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, got)
		}
	}
}

// TestStatementMatchesCondition checks condition-matching logic for statements
func TestStatementMatchesCondition(t *testing.T) {
	// FIXME(nsiow)
}

// TestEvalIsSameAccount checks same vs x-account checking behavior
func TestEvalIsSameAccount(t *testing.T) {
	type test struct {
		principal entities.Principal
		resource  entities.Resource
		want      bool
	}

	tests := []test{
		{
			principal: entities.Principal{Account: "88888"},
			resource:  entities.Resource{Account: "88888"},
			want:      true,
		},
		{
			principal: entities.Principal{Account: "88888"},
			resource:  entities.Resource{Account: "12345"},
			want:      false,
		},
	}

	for _, tc := range tests {
		got := evalIsSameAccount(&tc.principal, &tc.resource)
		if got != tc.want {
			t.Fatalf("failed, wanted %v got %v for test case: %v", tc.want, got, tc)
		}
	}
}
