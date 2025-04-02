package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/trace"
)

// TestStatementMatchesAction checks action-matching logic for statements
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesAction(trace.New(), &Options{}, i.ac, &i.stmt)
	})
}

// TestStatementMatchesPrincipal checks principal-matching logic for statements
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
			Want: true,
		},
		// Principal
		{
			Name: "simple_wildcard",
			Input: input{
				ac:   AuthContext{Principal: entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"*"}}},
			},
			Want: true,
		},
		{
			Name: "simple_direct_match",
			Input: input{
				ac:   AuthContext{Principal: entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerole"}}},
			},
			Want: true,
		},
		{
			Name: "other_principal",
			Input: input{
				ac:   AuthContext{Principal: entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerandomrole"}}},
			},
			Want: false,
		},
		{
			Name: "two_principals",
			Input: input{
				ac: AuthContext{Principal: entities.Principal{Arn: "arn:aws:iam::88888:role/secondrole"}},
				stmt: policy.Statement{Principal: policy.Principal{AWS: []string{
					"arn:aws:iam::88888:role/firstrole",
					"arn:aws:iam::88888:role/secondrole"}}}},
			Want: true,
		},
		{
			Name: "other_service",
			Input: input{
				ac:   AuthContext{Principal: entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{Principal: policy.Principal{Federated: []string{"*"}}},
			},
			Want: false,
		},
		{
			Name: "special_principal_star",
			Input: input{
				ac:   AuthContext{Principal: entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{Principal: policy.Principal{All: true}},
			},
			Want: true,
		},

		// NotPrincipal
		{
			Name: "notprincipal_simple_wildcard",
			Input: input{
				ac:   AuthContext{Principal: entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"*"}}},
			},
			Want: false,
		},
		{
			Name: "notprincipal_simple_direct_match",
			Input: input{
				ac:   AuthContext{Principal: entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerole"}}},
			},
			Want: false,
		},
		{
			Name: "notprincipal_other_principal",
			Input: input{
				ac:   AuthContext{Principal: entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"arn:aws:iam::88888:role/somerandomrole"}}},
			},
			Want: true,
		},
		{
			Name: "notprincipal_two_principals",
			Input: input{
				ac: AuthContext{Principal: entities.Principal{Arn: "arn:aws:iam::88888:role/secondrole"}},
				stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{
					"arn:aws:iam::88888:role/firstrole",
					"arn:aws:iam::88888:role/secondrole"}}}},
			Want: false,
		},
		{
			Name: "notprincipal_other_service",
			Input: input{
				ac:   AuthContext{Principal: entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{NotPrincipal: policy.Principal{Federated: []string{"*"}}},
			},
			Want: true,
		},
		{
			Name: "special_notprincipal_star",
			Input: input{
				ac:   AuthContext{Principal: entities.Principal{Arn: "arn:aws:iam::88888:role/somerole"}},
				stmt: policy.Statement{NotPrincipal: policy.Principal{All: true}},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesPrincipal(trace.New(), &Options{}, i.ac, &i.stmt)
	})
}

// TestStatementMatchesResource checks resource-matching logic for statements
func TestStatementMatchesResource(t *testing.T) {
	type input struct {
		ac   AuthContext
		stmt policy.Statement
	}

	tests := []testlib.TestCase[input, bool]{
		// Missing
		{
			Name: "missing_resource",
			Input: input{
				ac:   AuthContext{},
				stmt: policy.Statement{Resource: []string{"*"}},
			},
			Want: true,
		},
		// Resource
		{
			Name: "simple_wildcard",
			Input: input{
				ac:   AuthContext{Resource: entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
				stmt: policy.Statement{Resource: []string{"*"}},
			},
			Want: true,
		},
		{
			Name: "simple_direct_match",
			Input: input{
				ac:   AuthContext{Resource: entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
				stmt: policy.Statement{Resource: []string{"arn:aws:s3:::somebucket"}},
			},
			Want: true,
		},
		{
			Name: "other_resource",
			Input: input{
				ac:   AuthContext{Resource: entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
				stmt: policy.Statement{Resource: []string{"arn:aws:s3:::adifferentbucket"}},
			},
			Want: false,
		},
		{
			Name: "two_resources",
			Input: input{
				ac: AuthContext{Resource: entities.Resource{Arn: "arn:aws:s3:::secondbucket"}},
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
				ac:   AuthContext{Resource: entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
				stmt: policy.Statement{NotResource: []string{"*"}},
			},
			Want: false,
		},
		{
			Name: "notresource_simple_direct_match",
			Input: input{
				ac:   AuthContext{Resource: entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
				stmt: policy.Statement{NotResource: []string{"arn:aws:s3:::somebucket"}},
			},
			Want: false,
		},
		{
			Name: "notresource_other_resource",
			Input: input{
				ac:   AuthContext{Resource: entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
				stmt: policy.Statement{NotResource: []string{"arn:aws:s3:::adifferentbucket"}},
			},
			Want: true,
		},
		{
			Name: "notresource_two_resources",
			Input: input{
				ac: AuthContext{Resource: entities.Resource{Arn: "arn:aws:s3:::secondbucket"}},
				stmt: policy.Statement{NotResource: []string{
					"arn:aws:s3:::firstbucket",
					"arn:aws:s3:::secondbucket"}},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesResource(trace.New(), &Options{}, i.ac, &i.stmt)
	})
}
