package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestStatementMatchesAction(t *testing.T) {
	type input struct {
		ac   AuthContext
		stmt policy.Statement
	}

	tests := []testlib.TestCase[input, bool]{
		// Missing
		{
			Name: "missing_action",
			Input: input{
				ac:   AuthContext{},
				stmt: policy.Statement{Action: []string{"*"}},
			},
			Want: false,
		},
		// Action
		{
			Name: "simple_wildcard",
			Input: input{
				ac:   AuthContext{Action: sar.MustLookupString("s3:getobject")},
				stmt: policy.Statement{Action: []string{"*"}},
			},
			Want: true,
		},
		{
			Name: "simple_direct_match",
			Input: input{
				ac:   AuthContext{Action: sar.MustLookupString("s3:getobject")},
				stmt: policy.Statement{Action: []string{"s3:getobject"}},
			},
			Want: true,
		},
		{
			Name: "other_action",
			Input: input{
				ac:   AuthContext{Action: sar.MustLookupString("s3:putobject")},
				stmt: policy.Statement{Action: []string{"s3:getobject"}},
			},
			Want: false,
		},
		{
			Name: "two_actions",
			Input: input{
				ac:   AuthContext{Action: sar.MustLookupString("s3:getobject")},
				stmt: policy.Statement{Action: []string{"s3:putobject", "s3:getobject"}},
			},
			Want: true,
		},
		{
			Name: "diff_casing",
			Input: input{
				ac:   AuthContext{Action: sar.MustLookupString("s3:gEtObJeCt")},
				stmt: policy.Statement{Action: []string{"s3:putobject", "s3:getobject"}},
			},
			Want: true,
		},

		// NotAction
		{
			Name: "notaction_simple_wildcard",
			Input: input{
				ac:   AuthContext{Action: sar.MustLookupString("s3:getobject")},
				stmt: policy.Statement{NotAction: []string{"*"}},
			},
			Want: false,
		},
		{
			Name: "notaction_simple_direct_match",
			Input: input{
				ac:   AuthContext{Action: sar.MustLookupString("s3:getobject")},
				stmt: policy.Statement{NotAction: []string{"s3:getobject"}},
			},
			Want: false,
		},
		{
			Name: "notaction_other_action",
			Input: input{
				ac:   AuthContext{Action: sar.MustLookupString("sqs:sendmessage")},
				stmt: policy.Statement{NotAction: []string{"s3:getobject"}},
			},
			Want: true,
		},
		{
			Name: "notaction_two_actions",
			Input: input{
				ac:   AuthContext{Action: sar.MustLookupString("s3:getobject")},
				stmt: policy.Statement{NotAction: []string{"s3:putobject", "s3:getobject"}},
			},
			Want: false,
		},
		{
			Name: "notaction_diff_casing",
			Input: input{
				ac:   AuthContext{Action: sar.MustLookupString("s3:gEtObJeCt")},
				stmt: policy.Statement{NotAction: []string{"s3:putobject", "s3:getobject"}},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, Options{})
		return evalStatementMatchesAction(subj, &i.stmt), nil
	})
}

func TestStatementMatchesPrincipal(t *testing.T) {
	type input struct {
		ac   AuthContext
		stmt policy.Statement
	}

	tests := []testlib.TestCase[input, bool]{
		// Missing
		{
			Name: "missing_principal",
			Input: input{
				ac:   AuthContext{},
				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"*"}}},
			},
			Want: false,
		},
		// Principal
		{
			Name: "simple_wildcard",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"*"}}},
			},
			Want: true,
		},
		{
			Name: "simple_direct_match",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerole"}}},
			},
			Want: true,
		},
		{
			Name: "account_root_match_syntax_1",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"88888"}}},
			},
			Want: true,
		},
		{
			Name: "account_root_nomatch_syntax_1",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"55555"}}},
			},
			Want: false,
		},
		{
			Name: "account_root_match_syntax_2",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::88888:root"}}},
			},
			Want: true,
		},
		{
			Name: "account_root_nomatch_syntax_2",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::55555:root"}}},
			},
			Want: false,
		},
		{
			Name: "account_root_nomatch_syntax_3",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::55555:root/test"}}},
			},
			Want: false,
		},
		{
			Name: "other_principal",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerandomrole"}}},
			},
			Want: false,
		},
		{
			Name: "two_principals",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/secondrole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{
					"arn:aws:iam::88888:role/firstrole",
					"arn:aws:iam::88888:role/secondrole"}}}},
			Want: true,
		},
		{
			Name: "other_service",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{Federated: []string{"*"}}},
			},
			Want: false,
		},
		{
			Name: "special_principal_star",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{All: true}},
			},
			Want: true,
		},

		// NotPrincipal
		{
			Name: "notprincipal_simple_wildcard",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"*"}}},
			},
			Want: false,
		},
		{
			Name: "notprincipal_simple_direct_match",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerole"}}},
			},
			Want: false,
		},
		{
			Name: "notprincipal_other_principal",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerandomrole"}}},
			},
			Want: true,
		},
		{
			Name: "notprincipal_two_principals",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/secondrole",
					},
				},

				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{
					"arn:aws:iam::88888:role/firstrole",
					"arn:aws:iam::88888:role/secondrole"}}}},
			Want: false,
		},
		{
			Name: "notprincipal_other_service",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{NotPrincipal: policy.Principal{Federated: []string{"*"}}},
			},
			Want: true,
		},
		{
			Name: "special_notprincipal_star",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						AccountId: "88888",
						Arn:       "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{NotPrincipal: policy.Principal{All: true}},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesPrincipal(subj, &i.stmt), nil
	})
}

func TestStatementMatchesPrincipalExact(t *testing.T) {
	type input struct {
		ac   AuthContext
		stmt policy.Statement
	}

	tests := []testlib.TestCase[input, bool]{
		// Missing
		{
			Name: "missing_principal",
			Input: input{
				ac:   AuthContext{},
				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"*"}}},
			},
			Want: false,
		},
		// Principal
		{
			Name: "simple_direct_match",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Arn: "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerole"}}},
			},
			Want: true,
		},
		{
			Name: "other_principal",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Arn: "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerandomrole"}}},
			},
			Want: false,
		},
		{
			Name: "two_principals",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Arn: "arn:aws:iam::88888:role/secondrole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{
					"arn:aws:iam::88888:role/firstrole",
					"arn:aws:iam::88888:role/secondrole"}}}},
			Want: true,
		},
		{
			Name: "special_principal_star",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Arn: "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{Principal: policy.Principal{All: true}},
			},
			Want: false,
		},

		// NotPrincipal
		{
			Name: "notprincipal_simple_wildcard",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Arn: "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"*"}}},
			},
			Want: false,
		},
		{
			Name: "notprincipal_simple_direct_match",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Arn: "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerole"}}},
			},
			Want: false,
		},
		{
			Name: "notprincipal_other_principal",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Arn: "arn:aws:iam::88888:role/somerole",
					},
				},

				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerandomrole"}}},
			},
			Want: false,
		},
		{
			Name: "notprincipal_two_principals",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Arn: "arn:aws:iam::88888:role/secondrole",
					},
				},

				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{
					"arn:aws:iam::88888:role/firstrole",
					"arn:aws:iam::88888:role/secondrole"}}}},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesPrincipalExact(subj, &i.stmt), nil
	})
}

func TestStatementMatchesResource(t *testing.T) {
	type input struct {
		ac   AuthContext
		stmt policy.Statement
	}

	tests := []testlib.TestCase[input, bool]{
		// Missing
		{
			Name: "missing_resource_expected",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:listallmybuckets"),
				},
				stmt: policy.Statement{Resource: []string{"*"}},
			},
			Want: true,
		},
		// Missing
		{
			Name: "missing_resource_unexpected",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:listbucket"),
				},
				stmt: policy.Statement{Resource: []string{"*"}},
			},
			Want: false,
		},
		// Resource
		{
			Name: "simple_wildcard",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:listbucket"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:s3:::somebucket",
					},
				},

				stmt: policy.Statement{Resource: []string{"*"}},
			},
			Want: true,
		},
		{
			Name: "simple_direct_match",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:listbucket"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:s3:::somebucket",
					},
				},

				stmt: policy.Statement{Resource: []string{"arn:aws:s3:::somebucket"}},
			},
			Want: true,
		},
		{
			Name: "other_resource",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:listbucket"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:s3:::somebucket",
					},
				},

				stmt: policy.Statement{Resource: []string{"arn:aws:s3:::adifferentbucket"}},
			},
			Want: false,
		},
		{
			Name: "two_resources",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:listbucket"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:s3:::secondbucket",
					},
				},

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
				ac: AuthContext{
					Action: sar.MustLookupString("s3:listbucket"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:s3:::somebucket",
					},
				},

				stmt: policy.Statement{NotResource: []string{"*"}},
			},
			Want: false,
		},
		{
			Name: "notresource_simple_direct_match",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:listbucket"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:s3:::somebucket",
					},
				},

				stmt: policy.Statement{NotResource: []string{"arn:aws:s3:::somebucket"}},
			},
			Want: false,
		},
		{
			Name: "notresource_other_resource",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:listbucket"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:s3:::somebucket",
					},
				},

				stmt: policy.Statement{NotResource: []string{"arn:aws:s3:::adifferentbucket"}},
			},
			Want: true,
		},
		{
			Name: "notresource_two_resources",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:listbucket"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:s3:::secondbucket",
					},
				},

				stmt: policy.Statement{NotResource: []string{
					"arn:aws:s3:::firstbucket",
					"arn:aws:s3:::secondbucket"}},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, NewOptions())
		return evalStatementMatchesResource(subj, &i.stmt), nil
	})
}
