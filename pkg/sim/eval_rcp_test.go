package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestRCP(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, Decision]{
		{
			Name: "no_rcps",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::55555:role/myrole",
				},
				Resource: &entities.FrozenResource{
					Type: "AWS::S3::Bucket",
					Arn:  "arn:aws:s3:::mybucket",
					Account: entities.FrozenAccount{
						RCPs: [][]entities.ManagedPolicy{},
					},
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{Allow: true},
		},
		{
			Name: "no_resource",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::55555:role/myrole",
				},
				Action: sar.MustLookupString("s3:listallmybuckets"),
			},
			Want: Decision{Allow: true},
		},
		{
			Name: "rcp_unsupported_service",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::55555:role/myrole",
				},
				Resource: &entities.FrozenResource{
					Type: "AWS::SNS::Topic",
					Arn:  "arn:aws:sns:us-west-2:55555:mytopic",
					Account: entities.FrozenAccount{
						RCPs: [][]entities.ManagedPolicy{
							{
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
					},
				},
				Action: sar.MustLookupString("sns:publish"),
			},
			Want: Decision{Allow: true},
		},
		{
			Name: "rcp_role_unrelated_action",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::55555:role/myrole",
				},
				Resource: &entities.FrozenResource{
					Type: "AWS::IAM::Role",
					Arn:  "arn:aws:iam::55555:role/myrole-2",
					Account: entities.FrozenAccount{
						RCPs: [][]entities.ManagedPolicy{
							{
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
					},
				},
				Action: sar.MustLookupString("iam:getrole"),
			},
			Want: Decision{Allow: true},
		},
		{
			Name: "rcp_role_sts",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::55555:role/myrole",
				},
				Resource: &entities.FrozenResource{
					Type: "AWS::IAM::Role",
					Arn:  "arn:aws:iam::55555:role/myrole-2",
					Account: entities.FrozenAccount{
						RCPs: [][]entities.ManagedPolicy{
							{
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
					},
				},
				Action: sar.MustLookupString("sts:assumerole"),
			},
			Want: Decision{Deny: true},
		},
		{
			Name: "rcp_allow_all",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::55555:role/myrole",
				},
				Resource: &entities.FrozenResource{
					Type: "AWS::S3::Bucket",
					Arn:  "arn:aws:s3:::mybucket",
					Account: entities.FrozenAccount{
						RCPs: [][]entities.ManagedPolicy{
							{
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
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{Allow: true},
		},
		{
			Name: "rcp_deny_all",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::55555:role/myrole",
				},
				Resource: &entities.FrozenResource{
					Type: "AWS::S3::Bucket",
					Arn:  "arn:aws:s3:::mybucket",
					Account: entities.FrozenAccount{
						RCPs: [][]entities.ManagedPolicy{
							{
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
					},
				},
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{Deny: true},
		},
		{
			Name: "rcp_allowed_service",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::55555:role/myrole",
				},
				Resource: &entities.FrozenResource{
					Type: "AWS::S3::Bucket",
					Arn:  "arn:aws:s3:::mybucket",
					Account: entities.FrozenAccount{
						RCPs: [][]entities.ManagedPolicy{
							{
								{
									Policy: policy.Policy{
										Statement: []policy.Statement{
											{
												Effect: policy.EFFECT_ALLOW,
												Action: []string{"s3:*"},
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
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{Allow: true},
		},
		{
			Name: "rcp_not_allowed_service",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::55555:role/myrole",
				},
				Resource: &entities.FrozenResource{
					Type: "AWS::S3::Bucket",
					Arn:  "arn:aws:s3:::mybucket",
					Account: entities.FrozenAccount{
						RCPs: [][]entities.ManagedPolicy{
							{
								{
									Policy: policy.Policy{
										Statement: []policy.Statement{
											{
												Effect: policy.EFFECT_ALLOW,
												Action: []string{"ec2:*"},
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
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{},
		},
		{
			Name: "rcp_mid_layer_implicit_deny",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::55555:role/myrole",
				},
				Resource: &entities.FrozenResource{
					Type: "AWS::S3::Bucket",
					Arn:  "arn:aws:s3:::mybucket",
					Account: entities.FrozenAccount{
						RCPs: [][]entities.ManagedPolicy{
							{
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
							{}, // should cause a deny
							{
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
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{},
		},
		{
			Name: "rcp_mid_layer_explicit_deny",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{
					Arn: "arn:aws:iam::55555:role/myrole",
				},
				Resource: &entities.FrozenResource{
					Type: "AWS::S3::Bucket",
					Arn:  "arn:aws:s3:::mybucket",
					Account: entities.FrozenAccount{
						RCPs: [][]entities.ManagedPolicy{
							{
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
							{
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
							{
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
				Action: sar.MustLookupString("s3:ListBucket"),
			},
			Want: Decision{Deny: true},
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) (Decision, error) {
		subj := newSubject(&ac, TestingSimulationOptions)
		decision := evalRCP(subj)
		return decision, nil
	})
}
