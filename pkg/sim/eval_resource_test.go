package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestResourceAccess(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, Decision]{
		{
			Name: "implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: &entities.FrozenResource{
					Arn:    "arn:aws:s3:::mybucket",
					Policy: policy.Policy{},
				},
			},
			Want: Decision{},
		},
		{
			Name: "simple_match",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: &entities.FrozenResource{
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
			Want: Decision{Allow: true},
		},
		{
			Name: "explicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: &entities.FrozenResource{
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
			Want: Decision{Deny: true},
		},
		{
			Name: "allow_and_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::88888:role/myrole",
				},
				Resource: &entities.FrozenResource{
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
			Want: Decision{Allow: true, Deny: true},
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) (Decision, error) {
		subj := newSubject(&ac, TestingSimulationOptions)
		decision := evalResourceAccess(subj)
		return decision, nil
	})
}
