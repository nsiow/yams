package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestPrincipalAccess(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, []policy.Effect]{
		{
			Name: "implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Principal: entities.Principal{
						Arn:              "arn:aws:iam::88888:role/myrole",
						InlinePolicies:   nil,
						AttachedPolicies: nil,
					},
				},
				Resource: &entities.FrozenResource{
					Resource: entities.Resource{
						Arn: "arn:aws:s3:::mybucket",
					},
				},
			},
			Want: []policy.Effect(nil),
		},
		{
			Name: "simple_inline_policy",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Principal: entities.Principal{
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
				},
				Resource: &entities.FrozenResource{
					Resource: entities.Resource{
						Arn: "arn:aws:s3:::mybucket",
					},
				},
			},
			Want: []policy.Effect{policy.EFFECT_ALLOW},
		},
		{
			Name: "simple_named_policy",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Principal: entities.Principal{
						Arn: "arn:aws:iam::88888:role/myrole",
						InlinePolicies: []policy.Policy{
							{
								Id: "foo",
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
				},
				Resource: &entities.FrozenResource{
					Resource: entities.Resource{
						Arn: "arn:aws:s3:::mybucket",
					},
				},
			},
			Want: []policy.Effect{policy.EFFECT_ALLOW},
		},
		{
			Name: "simple_attached_policy",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Principal: entities.Principal{
						Arn: "arn:aws:iam::88888:role/myrole",
					},
					FrozenAttachedPolicies: []entities.ManagedPolicy{
						{
							Policy: policy.Policy{
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
				},
				Resource: &entities.FrozenResource{
					Resource: entities.Resource{
						Arn: "arn:aws:s3:::mybucket",
					},
				},
			},
			Want: []policy.Effect{policy.EFFECT_ALLOW},
		},
		{
			Name: "simple_group_policy",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Principal: entities.Principal{
						Arn: "arn:aws:iam::88888:role/myrole",
					},
					FrozenGroups: []entities.FrozenGroup{
						{
							Group: entities.Group{
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
							FrozenAttachedPolicies: []entities.ManagedPolicy{
								{
									Policy: policy.Policy{
										Statement: []policy.Statement{
											{
												Effect:   policy.EFFECT_ALLOW,
												Action:   []string{"ec2:describeinstances"},
												Resource: []string{"*"},
											},
										},
									},
								},
							},
						},
					},
					FrozenAttachedPolicies: []entities.ManagedPolicy{
						{
							Policy: policy.Policy{
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
				},
				Resource: &entities.FrozenResource{
					Resource: entities.Resource{
						Arn: "arn:aws:s3:::mybucket",
					},
				},
			},
			Want: []policy.Effect{policy.EFFECT_ALLOW},
		},
		{
			Name: "simple_inline_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Principal: entities.Principal{
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
				},
				Resource: &entities.FrozenResource{
					Resource: entities.Resource{
						Arn: "arn:aws:s3:::mybucket",
					},
				},
			},
			Want: []policy.Effect{policy.EFFECT_DENY},
		},
		{
			Name: "simple_attached_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Principal: entities.Principal{
						Arn: "arn:aws:iam::88888:role/myrole",
					},
					FrozenAttachedPolicies: []entities.ManagedPolicy{
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
					Resource: entities.Resource{
						Arn: "arn:aws:s3:::mybucket",
					},
				},
			},
			Want: []policy.Effect{policy.EFFECT_DENY},
		},
		{
			Name: "allow_and_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Principal: entities.Principal{
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
					FrozenAttachedPolicies: []entities.ManagedPolicy{
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
					Resource: entities.Resource{
						Arn: "arn:aws:s3:::mybucket",
					},
				},
			},
			Want: []policy.Effect{policy.EFFECT_ALLOW, policy.EFFECT_DENY},
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) ([]policy.Effect, error) {
		subj := newSubject(&ac, TestingSimulationOptions)
		decision := evalPrincipalAccess(subj)
		return decision.Effects(), nil
	})
}
