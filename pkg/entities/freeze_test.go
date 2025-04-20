package entities

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/policy"
)

func TestFreeze(t *testing.T) {
	type output struct {
		fp []FrozenPrincipal
		fr []FrozenResource
	}

	tests := []testlib.TestCase[*Universe, output]{
		{
			Name:  "empty_universe",
			Input: NewBuilder().Build(),
			Want:  output{},
		},
		{
			Name: "valid_single_principal",
			Input: NewBuilder().
				WithPrincipals(
					Principal{
						Arn: "arn:aws:iam::88888:role/role1",
					},
				).
				Build(),
			Want: output{
				fp: []FrozenPrincipal{
					{
						Principal: Principal{
							Arn: "arn:aws:iam::88888:role/role1",
						},
					},
				},
			},
		},
		{
			Name: "invalid_single_principal_permission_boundary",
			Input: NewBuilder().
				WithPrincipals(
					Principal{
						PermissionsBoundary: Arn("arn:aws:iam::aws:policy/AmazonS3FullAccess"),
					},
				).
				Build(),
			ShouldErr: true,
		},
		{
			Name: "invalid_single_principal_scp",
			Input: NewBuilder().
				WithPrincipals(
					Principal{
						AccountId: "55555",
					},
				).
				WithAccounts(
					Account{
						Id: "55555",
						SCPs: [][]Arn{
							{
								Arn("arn:aws:organizations::55555:policy/o-123/service_control_policy/p-123"),
							},
						},
					},
				).
				Build(),
			ShouldErr: true,
		},
		{
			Name: "invalid_single_principal_managed_policy",
			Input: NewBuilder().
				WithPrincipals(
					Principal{
						AttachedPolicies: []Arn{
							Arn("arn:aws:iam::aws:policy/AmazonS3FullAccess"),
						},
					},
				).
				Build(),
			ShouldErr: true,
		},
		{
			Name: "invalid_single_principal_group",
			Input: NewBuilder().
				WithPrincipals(
					Principal{
						Groups: []Arn{
							Arn("arn:aws:iam::55555:group/group-1"),
						},
					},
				).
				Build(),
			ShouldErr: true,
		},

		{
			Name: "invalid_single_principal_group_policy",
			Input: NewBuilder().
				WithPrincipals(
					Principal{
						Groups: []Arn{
							Arn("arn:aws:iam::55555:group/group-1"),
						},
					},
				).
				WithGroups(
					Group{
						Arn: "arn:aws:iam::55555:group/group-1",
						AttachedPolicies: []Arn{
							Arn("arn:aws:iam::55555:policy/p-123"),
						},
					},
				).
				Build(),
			ShouldErr: true,
		},
		{
			Name: "valid_single_resource",
			Input: NewBuilder().
				WithResources(
					Resource{
						Arn: "arn:aws:s3:::some-bucket",
					},
				).
				Build(),
			Want: output{
				fr: []FrozenResource{
					{
						Resource: Resource{
							Arn: "arn:aws:s3:::some-bucket",
						},
					},
				},
			},
		},
		{
			Name: "invalid_single_resource",
			Input: NewBuilder().
				WithResources(
					Resource{
						AccountId: "55555",
					},
				).
				WithAccounts(
					Account{
						Id: "55555",
						SCPs: [][]Arn{
							{
								Arn("arn:aws:organizations::55555:policy/o-123/service_control_policy/p-123"),
							},
						},
					},
				).
				Build(),
			ShouldErr: true,
		},
		{
			Name: "valid_principal_and_account",
			Input: NewBuilder().
				WithPrincipals(
					Principal{
						AccountId: "55555",
						Arn:       "arn:aws:iam::55555:role/role1",
					},
				).
				WithAccounts(
					Account{
						Id: "55555",
						SCPs: [][]Arn{
							{
								Arn("arn:aws:organizations::55555:policy/o-123/service_control_policy/p-123"),
							},
						},
					},
				).
				WithPolicies(
					ManagedPolicy{
						Arn: "arn:aws:organizations::55555:policy/o-123/service_control_policy/p-123",
						Policy: policy.Policy{
							Statement: policy.StatementBlock{
								{
									Sid: "stmt0",
								},
							},
						},
					},
				).
				Build(),
			Want: output{
				fp: []FrozenPrincipal{
					{
						Principal: Principal{
							AccountId: "55555",
							Arn:       "arn:aws:iam::55555:role/role1",
						},
						FrozenAccount: FrozenAccount{
							Account: Account{
								Id: "55555",
								SCPs: [][]Arn{
									{
										Arn("arn:aws:organizations::55555:policy/o-123/service_control_policy/p-123"),
									},
								},
							},
							FrozenSCPs: [][]ManagedPolicy{
								{
									ManagedPolicy{
										Arn: "arn:aws:organizations::55555:policy/o-123/service_control_policy/p-123",
										Policy: policy.Policy{
											Statement: policy.StatementBlock{
												{
													Sid: "stmt0",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},

		{
			Name: "valid_principal_and_group",
			Input: NewBuilder().
				WithPrincipals(
					Principal{
						AccountId: "55555",
						Arn:       "arn:aws:iam::55555:role/role1",
						Groups: []Arn{
							Arn("arn:aws:iam::55555:group/group-1"),
						},
					},
				).
				WithGroups(
					Group{
						Arn: "arn:aws:iam::55555:group/group-1",
						AttachedPolicies: []Arn{
							Arn("arn:aws:iam::55555:policy/p-123"),
						},
					},
				).
				WithPolicies(
					ManagedPolicy{
						Arn: "arn:aws:iam::55555:policy/p-123",
						Policy: policy.Policy{
							Statement: policy.StatementBlock{
								{
									Sid: "stmt0",
								},
							},
						},
					},
				).
				Build(),
			Want: output{
				fp: []FrozenPrincipal{
					{
						Principal: Principal{
							AccountId: "55555",
							Arn:       "arn:aws:iam::55555:role/role1",
							Groups: []Arn{
								Arn("arn:aws:iam::55555:group/group-1"),
							},
						},
						FrozenGroups: []FrozenGroup{
							{
								Group: Group{
									Arn: "arn:aws:iam::55555:group/group-1",
									AttachedPolicies: []Arn{
										Arn("arn:aws:iam::55555:policy/p-123"),
									},
								},
								FrozenAttachedPolicies: []ManagedPolicy{
									{
										Arn: "arn:aws:iam::55555:policy/p-123",
										Policy: policy.Policy{
											Statement: policy.StatementBlock{
												{
													Sid: "stmt0",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	testlib.RunTestSuite(t, tests, func(input *Universe) (output, error) {
		out := output{}

		var err error

		out.fp, err = input.FrozenPrincipals()
		if err != nil {
			return output{}, err
		}

		out.fr, err = input.FrozenResources()
		if err != nil {
			return output{}, err
		}

		// remove universe association; no way to specify this in tests, so gross hack for now
		for i := range out.fp {
			out.fp[i].uv = nil
			out.fp[i].FrozenAccount.uv = nil
			for j := range out.fp[i].FrozenGroups {
				out.fp[i].FrozenGroups[j].uv = nil
				for k := range out.fp[i].FrozenGroups[j].FrozenAttachedPolicies {
					out.fp[i].FrozenGroups[j].FrozenAttachedPolicies[k].uv = nil
				}
			}
			for j := range out.fp[i].FrozenAccount.FrozenSCPs {
				for k := range out.fp[i].FrozenAccount.FrozenSCPs[j] {
					out.fp[i].FrozenAccount.FrozenSCPs[j][k].uv = nil
				}
			}
		}
		for i := range out.fr {
			out.fr[i].uv = nil
			out.fr[i].FrozenAccount.uv = nil
			for j := range out.fr[i].FrozenAccount.FrozenSCPs {
				for k := range out.fr[i].FrozenAccount.FrozenSCPs[j] {
					out.fr[i].FrozenAccount.FrozenSCPs[j][k].uv = nil
				}
			}
		}

		return out, nil
	})
}

func TestFreeze_MissingUniverse(t *testing.T) {
	_, err := (&Account{}).Freeze()
	if err == nil {
		t.Fatalf("should have failed on account with missing universe")
	}

	_, err = (&Principal{}).Freeze()
	if err == nil {
		t.Fatalf("should have failed on principal with missing universe")
	}

	_, err = (&Resource{}).Freeze()
	if err == nil {
		t.Fatalf("should have failed on resource with missing universe")
	}

	_, err = (&Group{}).Freeze()
	if err == nil {
		t.Fatalf("should have failed on group with missing universe")
	}
}
