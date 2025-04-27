package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestPermissionsBoundary(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, Decision]{
		{
			Name: "allow_all",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					PermissionBoundary: entities.ManagedPolicy{
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
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{Allow: true},
		},
		{
			Name: "deny_all",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
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
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{Deny: true},
		},
		{
			Name: "allow_others_simple",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					PermissionBoundary: entities.ManagedPolicy{
						Policy: policy.Policy{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"ec2:DescribeInstances"},
									Resource: []string{"*"},
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
			Name: "allow_specific",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					PermissionBoundary: entities.ManagedPolicy{
						Policy: policy.Policy{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:ListBucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
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
			Name: "allow_others_specific",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					PermissionBoundary: entities.ManagedPolicy{
						Policy: policy.Policy{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:ListBucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
									Condition: map[string]map[string]policy.Value{
										"StringEquals": {
											"aws:UserAgent": []string{"some-random-ua"},
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
			Name: "allow_only_iam",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					PermissionBoundary: entities.ManagedPolicy{
						Policy: policy.Policy{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"iam:*"},
									Resource: []string{"*"},
								},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("iam:ListRoles"),
			},
			Want: Decision{Allow: true},
		},
		{
			Name: "deny_iam_by_omission",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
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
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("iam:ListRoles"),
			},
			Want: Decision{},
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) (Decision, error) {
		subj := newSubject(&ac, TestingSimulationOptions)
		decision := evalPermissionsBoundary(subj)
		return decision, nil
	})
}
