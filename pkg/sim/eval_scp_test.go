package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestSCP(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, Decision]{
		{
			Name: "no_scps",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{},
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{allow: true},
		},
		{
			Name: "scp_allow_all",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
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
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{allow: true},
		},
		{
			Name: "scp_deny_all",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
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
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{deny: true},
		},
		{
			Name: "scp_allowed_service",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
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
					},
				},
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{allow: true},
		},
		{
			Name: "scp_not_allowed_service",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
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
			Name: "scp_mid_layer_implicit_deny",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect: policy.EFFECT_ALLOW,
													Action: []string{"*"},
													Principal: policy.Principal{
														AWS: []string{"*"},
													},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
							{
								SCPs: []entities.ManagedPolicy{},
							},
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect: policy.EFFECT_ALLOW,
													Action: []string{"*"},
													Principal: policy.Principal{
														AWS: []string{"*"},
													},
													Resource: []string{"*"},
												},
											},
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
			Name: "scp_mid_layer_explicit_deny",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect: policy.EFFECT_ALLOW,
													Action: []string{"*"},
													Principal: policy.Principal{
														AWS: []string{"*"},
													},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect: policy.EFFECT_DENY,
													Action: []string{"*"},
													Principal: policy.Principal{
														AWS: []string{"*"},
													},
													Resource: []string{"*"},
												},
											},
										},
									},
								},
							},
							{
								SCPs: []entities.ManagedPolicy{
									{
										Policy: policy.Policy{
											Statement: []policy.Statement{
												{
													Effect: policy.EFFECT_ALLOW,
													Action: []string{"*"},
													Principal: policy.Principal{
														AWS: []string{"*"},
													},
													Resource: []string{"*"},
												},
											},
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
			Want: Decision{deny: true},
		},
		{
			Name: "scp_management_account_principal_exempt",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					AccountId: "99999",
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
										AccountId: "99999",
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
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{allow: true},
		},
		{
			Name: "scp_service_linked_role_exempt",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Type:      "AWS::IAM::Role",
					Arn:       "arn:aws:iam::55555:role/aws-service-role/s3.amazonaws.com/AWSServiceRoleForS3",
					AccountId: "55555",
					Account: entities.FrozenAccount{
						OrgNodes: []entities.FrozenOrgNode{
							{
								SCPs: []entities.ManagedPolicy{
									{
										AccountId: "99999",
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
							},
						},
					},
				},
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::mybucket",
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{allow: true},
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) (Decision, error) {
		subj := newSubject(ac, TestingSimulationOptions)
		decision := evalSCP(&subj)
		return decision, nil
	})
}
