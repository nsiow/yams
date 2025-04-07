package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestPermissionsBoundary(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, []policy.Effect]{
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
				Action:   sar.MustLookupString("s3:ListBucket"),
			},
			Want: []policy.Effect{
				policy.EFFECT_ALLOW,
			},
		},
		{
			Name: "deny_all",
			Input: AuthContext{
				Principal: &entities.Principal{
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
				Resource: &entities.Resource{Arn: "arn:aws:s3:::mybucket"},
				Action:   sar.MustLookupString("s3:ListBucket"),
			},
			Want: []policy.Effect{
				policy.EFFECT_DENY,
			},
		},
		{
			Name: "allow_others_simple",
			Input: AuthContext{
				Principal: &entities.Principal{
					PermissionsBoundary: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_ALLOW,
								Action:   []string{"ec2:DescribeInstances"},
								Resource: []string{"*"},
							},
						},
					},
				},
				Resource: &entities.Resource{Arn: "arn:aws:s3:::mybucket"},
				Action:   sar.MustLookupString("s3:ListBucket"),
			},
			Want: []policy.Effect(nil),
		},
		{
			Name: "allow_this_specific",
			Input: AuthContext{
				Principal: &entities.Principal{
					PermissionsBoundary: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_ALLOW,
								Action:   []string{"s3:ListBucket"},
								Resource: []string{"arn:aws:s3:::mybucket"},
							},
						},
					},
				},
				Resource: &entities.Resource{Arn: "arn:aws:s3:::mybucket"},
				Action:   sar.MustLookupString("s3:ListBucket"),
			},
			Want: []policy.Effect{
				policy.EFFECT_ALLOW,
			},
		},
		{
			Name: "allow_others_specific",
			Input: AuthContext{
				Principal: &entities.Principal{
					PermissionsBoundary: policy.Policy{
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
				Resource: &entities.Resource{Arn: "arn:aws:s3:::mybucket"},
				Action:   sar.MustLookupString("s3:ListBucket"),
			},
			Want: []policy.Effect(nil),
		},
		{
			Name: "allow_only_iam",
			Input: AuthContext{
				Principal: &entities.Principal{
					PermissionsBoundary: policy.Policy{
						Statement: []policy.Statement{
							{
								Effect:   policy.EFFECT_ALLOW,
								Action:   []string{"iam:*"},
								Resource: []string{"*"},
							},
						},
					},
				},
				Resource: &entities.Resource{Arn: "*"},
				Action:   sar.MustLookupString("iam:ListRoles"),
			},
			Want: []policy.Effect{
				policy.EFFECT_ALLOW,
			},
		},
		{
			Name: "deny_iam_by_omission",
			Input: AuthContext{
				Principal: &entities.Principal{
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
				Resource: &entities.Resource{Arn: "*"},
				Action:   sar.MustLookupString("iam:ListRoles"),
			},
			Want: []policy.Effect(nil),
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) ([]policy.Effect, error) {
		subj := newSubject(&ac, TestingSimulationOptions)
		res, err := evalPermissionsBoundary(subj)
		if err != nil {
			return nil, err
		}

		return res.Effects(), nil
	})
}
