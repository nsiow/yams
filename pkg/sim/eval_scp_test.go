package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/trace"
)

// TestSCP tests functionality of SCP evaluations
func TestSCP(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, []policy.Effect]{
		{
			Name: "no_scps",
			Input: AuthContext{
				Principal: &entities.Principal{},
				Resource:  &entities.Resource{Arn: "arn:aws:s3:::mybucket"},
				Action:    "s3:ListBucket",
			},
			Want: []policy.Effect{
				policy.EFFECT_ALLOW,
			},
		},
		{
			Name: "allow_all",
			Input: AuthContext{
				Principal: &entities.Principal{
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
				Resource: &entities.Resource{Arn: "arn:aws:s3:::mybucket"},
				Action:   "s3:ListBucket",
			},
			Want: []policy.Effect{
				policy.EFFECT_ALLOW,
			},
		},
		{
			Name: "deny_all",
			Input: AuthContext{
				Principal: &entities.Principal{
					Account: entities.Account{
						SCPs: [][]policy.Policy{
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
				Resource: &entities.Resource{Arn: "arn:aws:s3:::mybucket"},
				Action:   "s3:ListBucket",
			},
			Want: []policy.Effect{
				policy.EFFECT_DENY,
			},
		},
		{
			Name: "allowed_service",
			Input: AuthContext{
				Principal: &entities.Principal{
					Account: entities.Account{
						SCPs: [][]policy.Policy{
							{
								{
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
				Resource: &entities.Resource{Arn: "arn:aws:s3:::mybucket"},
				Action:   "s3:ListBucket",
			},
			Want: []policy.Effect{
				policy.EFFECT_ALLOW,
			},
		},
		{
			Name: "not_allowed_service",
			Input: AuthContext{
				Principal: &entities.Principal{
					Account: entities.Account{
						SCPs: [][]policy.Policy{
							{
								{
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
				Resource: &entities.Resource{Arn: "arn:aws:s3:::mybucket"},
				Action:   "s3:ListBucket",
			},
			Want: []policy.Effect(nil),
		},
		{
			Name: "mid_layer_implicit_deny",
			Input: AuthContext{
				Principal: &entities.Principal{
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
							{}, // <= should cause a deny
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
				Resource: &entities.Resource{Arn: "arn:aws:s3:::mybucket"},
				Action:   "s3:ListBucket",
			},
			Want: []policy.Effect(nil),
		},
		{
			Name: "mid_layer_explicit_deny",
			Input: AuthContext{
				Principal: &entities.Principal{
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
				Resource: &entities.Resource{Arn: "arn:aws:s3:::mybucket"},
				Action:   "s3:ListBucket",
			},
			Want: []policy.Effect{
				policy.EFFECT_DENY,
			},
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) ([]policy.Effect, error) {
		res, err := evalSCP(trace.New(), &Options{}, ac)
		if err != nil {
			return nil, err
		}

		return res.Effects(), nil
	})
}
