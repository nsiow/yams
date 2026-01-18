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
						Arn: "arn:aws:iam::88888:role/role1",
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
						OrgNodes: []OrgNode{
							{
								SCPs: []Arn{
									Arn("arn:aws:organizations::55555:policy/o-123/service_control_policy/p-123"),
								},
							},
						},
					},
				).
				Build(),
			ShouldErr: true,
		},
		{
			Name: "invalid_single_principal_rcp",
			Input: NewBuilder().
				WithPrincipals(
					Principal{
						AccountId: "55555",
					},
				).
				WithAccounts(
					Account{
						Id: "55555",
						OrgNodes: []OrgNode{
							{
								RCPs: []Arn{
									Arn("arn:aws:organizations::55555:policy/o-123/resource_control_policy/p-456"),
								},
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
						Arn: "arn:aws:s3:::some-bucket",
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
						OrgNodes: []OrgNode{
							{
								SCPs: []Arn{
									Arn("arn:aws:organizations::55555:policy/o-123/service_control_policy/p-123"),
								},
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
						OrgNodes: []OrgNode{
							{
								SCPs: []Arn{
									Arn("arn:aws:organizations::55555:policy/o-123/service_control_policy/p-123"),
								},
								RCPs: []Arn{
									Arn("arn:aws:organizations::55555:policy/o-123/resource_control_policy/p-456"),
								},
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
					ManagedPolicy{
						Arn: "arn:aws:organizations::55555:policy/o-123/resource_control_policy/p-456",
						Policy: policy.Policy{
							Statement: policy.StatementBlock{
								{
									Sid: "stmt1",
								},
							},
						},
					},
				).
				Build(),
			Want: output{
				fp: []FrozenPrincipal{
					{
						AccountId: "55555",
						Arn:       "arn:aws:iam::55555:role/role1",
						Account: FrozenAccount{
							Id: "55555",
							OrgNodes: []FrozenOrgNode{
								{
									SCPs: []ManagedPolicy{
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
									RCPs: []ManagedPolicy{
										ManagedPolicy{
											Arn: "arn:aws:organizations::55555:policy/o-123/resource_control_policy/p-456",
											Policy: policy.Policy{
												Statement: policy.StatementBlock{
													{
														Sid: "stmt1",
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
						AccountId: "55555",
						Arn:       "arn:aws:iam::55555:role/role1",
						Groups: []FrozenGroup{
							{
								Arn: "arn:aws:iam::55555:group/group-1",
								AttachedPolicies: []ManagedPolicy{
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

		out.fp, err = input.FrozenPrincipals(true, nil)
		if err != nil {
			return output{}, err
		}

		out.fr, err = input.FrozenResources(true, nil)
		if err != nil {
			return output{}, err
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

func TestFreeze_NonStrict(t *testing.T) {
	// Test non-strict freeze where policies and groups are missing - should create empty versions
	uv := NewBuilder().
		WithPrincipals(
			Principal{
				Arn:       "arn:aws:iam::88888:user/testuser",
				AccountId: "88888",
				AttachedPolicies: []Arn{
					"arn:aws:iam::88888:policy/missing-policy",
				},
				Groups: []Arn{
					"arn:aws:iam::88888:group/missing-group",
				},
				PermissionsBoundary: "arn:aws:iam::88888:policy/missing-boundary",
			},
		).
		WithAccounts(
			Account{
				Id: "88888",
				OrgNodes: []OrgNode{
					{
						SCPs: []Arn{
							"arn:aws:organizations::88888:policy/o-123/service_control_policy/missing-scp",
						},
						RCPs: []Arn{
							"arn:aws:organizations::88888:policy/o-123/resource_control_policy/missing-rcp",
						},
					},
				},
			},
		).
		Build()

	// Non-strict freeze should succeed and create empty policies/groups
	fps, err := uv.FrozenPrincipals(false, nil)
	if err != nil {
		t.Fatalf("non-strict freeze should not fail: %v", err)
	}

	if len(fps) != 1 {
		t.Fatalf("expected 1 frozen principal, got %d", len(fps))
	}

	fp := fps[0]

	// Verify attached policies were created as empty
	if len(fp.AttachedPolicies) != 1 {
		t.Fatalf("expected 1 attached policy, got %d", len(fp.AttachedPolicies))
	}
	if fp.AttachedPolicies[0].Arn != "arn:aws:iam::88888:policy/missing-policy" {
		t.Errorf("expected policy ARN to be preserved")
	}

	// Verify groups were created as empty
	if len(fp.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(fp.Groups))
	}
	if fp.Groups[0].Arn != "arn:aws:iam::88888:group/missing-group" {
		t.Errorf("expected group ARN to be preserved")
	}

	// Verify permission boundary was created as empty
	if fp.PermissionBoundary.Arn != "arn:aws:iam::88888:policy/missing-boundary" {
		t.Errorf("expected permission boundary ARN to be preserved")
	}

	// Verify account SCPs/RCPs were created as empty
	if len(fp.Account.OrgNodes) != 1 {
		t.Fatalf("expected 1 org node, got %d", len(fp.Account.OrgNodes))
	}
	if len(fp.Account.OrgNodes[0].SCPs) != 1 {
		t.Fatalf("expected 1 SCP, got %d", len(fp.Account.OrgNodes[0].SCPs))
	}
	if len(fp.Account.OrgNodes[0].RCPs) != 1 {
		t.Fatalf("expected 1 RCP, got %d", len(fp.Account.OrgNodes[0].RCPs))
	}
}

func TestFreeze_NonStrictResources(t *testing.T) {
	// Test non-strict freeze for resources with missing account policies
	uv := NewBuilder().
		WithResources(
			Resource{
				Arn:       "arn:aws:s3:::mybucket",
				AccountId: "88888",
			},
		).
		WithAccounts(
			Account{
				Id: "88888",
				OrgNodes: []OrgNode{
					{
						SCPs: []Arn{
							"arn:aws:organizations::88888:policy/o-123/service_control_policy/missing-scp",
						},
					},
				},
			},
		).
		Build()

	// Non-strict freeze should succeed
	frs, err := uv.FrozenResources(false, nil)
	if err != nil {
		t.Fatalf("non-strict freeze should not fail: %v", err)
	}

	if len(frs) != 1 {
		t.Fatalf("expected 1 frozen resource, got %d", len(frs))
	}
}
