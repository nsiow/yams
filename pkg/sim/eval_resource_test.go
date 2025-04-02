package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/trace"
)

// TestResourceAccess checks resource-policy evaluation logic for statements
func TestResourceAccess(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, []policy.Effect]{
		{
			Name: "implicit_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: entities.Resource{
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
				Principal: entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: entities.Resource{
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
				Principal: entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: entities.Resource{
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
				Principal: entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: entities.Resource{
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
				Principal: entities.Principal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: entities.Resource{
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

	testlib.RunTestSuite(t, tests, func(ac AuthContext) ([]policy.Effect, error) {
		opts := Options{FailOnUnknownCondition: true}
		res, err := evalResourceAccess(trace.New(), &opts, ac)
		if err != nil {
			return nil, err
		}

		return res.Effects(), nil
	})
}
