package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestSCP(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, Decision]{
		{
			Name: "no_scps",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						SCPs: [][]entities.ManagedPolicy{},
					},
				},
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{Allow: true},
		},
		{
			Name: "scp_allow_all",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						SCPs: [][]entities.ManagedPolicy{
							{
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
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{Allow: true},
		},
		{
			Name: "scp_deny_all",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						SCPs: [][]entities.ManagedPolicy{
							{
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
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{Deny: true},
		},
		{
			Name: "scp_allowed_service",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						SCPs: [][]entities.ManagedPolicy{
							{
								{
									Policy: policy.Policy{
										Statement: []policy.Statement{
											{
												Effect:   policy.EFFECT_ALLOW,
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
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{Allow: true},
		},
		{
			Name: "scp_not_allowed_service",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						SCPs: [][]entities.ManagedPolicy{
							{
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
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{},
		},
		{
			Name: "scp_mid_layer_implicit_deny",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						SCPs: [][]entities.ManagedPolicy{
							{
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
							{}, // should cause a deny
							{
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
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{},
		},
		{
			Name: "scp_mid_layer_explicit_deny",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						SCPs: [][]entities.ManagedPolicy{
							{
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
							{
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
							{
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
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{Deny: true},
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) (Decision, error) {
		subj := newSubject(&ac, TestingSimulationOptions)
		decision := evalSCP(subj)
		return decision, nil
	})
}
