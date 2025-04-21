package entities

import (
	"reflect"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestBuilder(t *testing.T) {
	type input struct {
		accounts   []Account
		groups     []Group
		policies   []ManagedPolicy
		principals []Principal
		resources  []Resource
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Name:  "empty",
			Input: input{},
			Want:  true,
		},
		{
			Name: "single_account",
			Input: input{
				accounts: []Account{
					{
						Id: "55555",
					},
				},
			},
			Want: true,
		},
		{
			Name: "many_things",
			Input: input{
				accounts: []Account{
					{
						Id: "55555",
					},
					{
						Id: "88888",
					},
				},
				groups: []Group{
					{
						Arn: "arn:aws:iam::55555:group/group-1",
					},
				},
				policies: []ManagedPolicy{
					{
						Arn: "arn:aws:iam::55555:policy/p-123",
					},
				},
				principals: []Principal{
					{
						Arn: "arn:aws:iam::55555:user/user-123",
					},
					{
						Arn: "arn:aws:iam::55555:role/role-123",
					},
					{
						Arn: "arn:aws:iam::55555:role/role-456",
					},
				},
				resources: []Resource{
					{
						Arn: "arn:aws:s3:::my-bucket",
					},
				},
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		uv_1 := NewUniverse()
		uv_2 := NewBuilder()

		for _, a := range i.accounts {
			uv_1.PutAccount(a)
			uv_2.WithAccounts(a)
		}

		for _, g := range i.groups {
			uv_1.PutGroup(g)
			uv_2.WithGroups(g)
		}

		for _, p := range i.policies {
			uv_1.PutPolicy(p)
			uv_2.WithPolicies(p)
		}

		for _, p := range i.principals {
			uv_1.PutPrincipal(p)
			uv_2.WithPrincipals(p)
		}

		for _, r := range i.resources {
			uv_1.PutResource(r)
			uv_2.WithResources(r)
		}

		return reflect.DeepEqual(uv_1, uv_2.Build()), nil
	})

}
