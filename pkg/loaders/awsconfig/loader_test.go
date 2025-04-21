package awsconfig

import (
	"os"
	"path"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestLoad(t *testing.T) {
	tests := []testlib.TestCase[string, *entities.Universe]{

		// ---------------------------------------------------------------------------------------------
		// Valid
		// ---------------------------------------------------------------------------------------------

		{
			Name:  "base_valid_empty",
			Input: `../../../testdata/config_loading/base_valid_empty.json`,
			Want:  entities.NewUniverse(),
		},
		{
			Name:  "base_valid_empty_json_l",
			Input: `../../../testdata/config_loading/base_valid_empty.jsonl`,
			Want:  entities.NewUniverse(),
		},
		{
			Name:  "generic_resource_valid.json",
			Input: `../../../testdata/config_loading/generic_resource_valid.json`,
			Want: entities.NewBuilder().
				WithResources(
					entities.Resource{
						Type:      "AWS::Lambda::Function",
						AccountId: "123456789012",
						Region:    "us-west-2",
						Arn:       "arn:aws:lambda:us-west-2:123456789012:function:my-function",
						Tags:      []entities.Tag{},
					},
				).
				Build(),
		},
		{
			Name:  "account_valid",
			Input: `../../../testdata/config_loading/account_valid.json`,
			Want: entities.NewBuilder().
				WithAccounts(
					entities.Account{
						Id:    "000000000000",
						OrgId: "o-123",
						OrgPaths: []string{
							"o-123/",
							"o-123/ou-level-1/",
							"o-123/ou-level-1/ou-level-2/",
						},
						SCPs: [][]entities.Arn{
							{
								"arn:aws:organizations::aws:policy/service_control_policy/p-FullAWSAccess/FullAWSAccess",
							},
							{
								"arn:aws:organizations::000000000000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access",
							},
						},
					},
				).
				Build(),
		},
		{
			Name:  "account_valid_jsonl",
			Input: `../../../testdata/config_loading/account_valid.jsonl`,
			Want: entities.NewBuilder().
				WithAccounts(
					entities.Account{
						Id:    "000000000000",
						OrgId: "o-123",
						OrgPaths: []string{
							"o-123/",
							"o-123/ou-level-1/",
							"o-123/ou-level-1/ou-level-2/",
						},
						SCPs: [][]entities.Arn{
							{
								"arn:aws:organizations::aws:policy/service_control_policy/p-FullAWSAccess/FullAWSAccess",
							},
							{
								"arn:aws:organizations::000000000000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access",
							},
						},
					},
				).
				Build(),
		},
		{
			Name:  "scp_valid",
			Input: `../../../testdata/config_loading/scp_valid.json`,
			Want: entities.NewBuilder().
				WithPolicies(
					entities.ManagedPolicy{
						Type:      "Yams::Organizations::ServiceControlPolicy",
						AccountId: "000000000000",
						Arn:       "arn:aws:organizations::000000000000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access",
						Policy: policy.Policy{
							Version: "2012-10-17",
							Statement: policy.StatementBlock{
								policy.Statement{
									Effect: "Allow",
									Action: policy.Value{
										"s3:*",
									},
									Resource: policy.Value{
										"*",
									},
								},
							},
						},
					},
				).
				WithResources(
					entities.Resource{
						Type:      "Yams::Organizations::ServiceControlPolicy",
						AccountId: "000000000000",
						Arn:       "arn:aws:organizations::000000000000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access",
					},
				).
				Build(),
		},
		{
			Name:  "group_valid",
			Input: `../../../testdata/config_loading/group_valid.json`,
			Want: entities.NewBuilder().
				WithGroups(
					entities.Group{
						Type:      "AWS::IAM::Group",
						AccountId: "000000000000",
						Arn:       "arn:aws:iam::000000000000:group/family",
						AttachedPolicies: []entities.Arn{
							"arn:aws:iam::000000000000:policy/Common",
						},
						InlinePolicies: []policy.Policy{
							{
								Version: "2012-10-17",
								Statement: policy.StatementBlock{
									policy.Statement{
										Effect: "Allow",
										Action: policy.Value{
											"s3:*",
										},
										Resource: policy.Value{
											"*",
										},
									},
								},
							},
						},
					},
				).
				WithResources(
					entities.Resource{
						Type:      "AWS::IAM::Group",
						AccountId: "000000000000",
						Arn:       "arn:aws:iam::000000000000:group/family",
						Tags:      []entities.Tag{},
					},
				).
				Build(),
		},
		{
			Name:  "policy_valid",
			Input: `../../../testdata/config_loading/policy_valid.json`,
			Want: entities.NewBuilder().
				WithPolicies(
					entities.ManagedPolicy{
						Type:      "AWS::IAM::Policy",
						AccountId: "000000000000",
						Arn:       "arn:aws:iam::000000000000:policy/Common",
						Policy: policy.Policy{
							Version: "2012-10-17",
							Statement: policy.StatementBlock{
								policy.Statement{
									Effect: "Allow",
									Action: policy.Value{
										"sqs:ReceiveMessage",
									},
									Resource: policy.Value{
										"arn:aws:sqs:us-east-1:0000000000000:queue-2",
									},
								},
							},
						},
					},
				).
				WithResources(
					entities.Resource{
						Type:      "AWS::IAM::Policy",
						AccountId: "000000000000",
						Arn:       "arn:aws:iam::000000000000:policy/Common",
						Tags:      []entities.Tag{},
					},
				).
				Build(),
		},
		{
			Name:  "role_valid",
			Input: `../../../testdata/config_loading/role_valid.json`,
			Want: entities.NewBuilder().
				WithPrincipals(
					entities.Principal{
						Type:      "AWS::IAM::Role",
						AccountId: "000000000000",
						Arn:       "arn:aws:iam::000000000000:role/SimpleRole1",
						Tags: []entities.Tag{
							{
								Key:   "some-business-tag",
								Value: "important-business-thing",
							},
							{
								Key:   "some-technical-tag",
								Value: "important-technical-thing",
							},
						},
						InlinePolicies: []policy.Policy{
							{
								Version: "2012-10-17",
								Statement: []policy.Statement{
									{
										Effect: "Allow",
										Action: []string{"s3:GetObject", "s3:ListBucket"},
										Resource: []string{
											"arn:aws:s3:::simple-bucket",
											"arn:aws:s3:::simple-bucket/*",
										},
									},
								},
							},
						},
						AttachedPolicies: []entities.Arn{
							"arn:aws:iam::000000000000:policy/Common",
						},
					},
				).
				WithResources(
					entities.Resource{
						Type:      "AWS::IAM::Role",
						AccountId: "000000000000",

						Arn: "arn:aws:iam::000000000000:role/SimpleRole1",
						Tags: []entities.Tag{
							{
								Key:   "some-business-tag",
								Value: "important-business-thing",
							},
							{
								Key:   "some-technical-tag",
								Value: "important-technical-thing",
							},
						},
						Policy: policy.Policy{
							Version: "2012-10-17",
							Statement: policy.StatementBlock{
								policy.Statement{
									Effect: "Allow",
									Principal: policy.Principal{
										Service: policy.Value{
											"ec2.amazonaws.com",
										},
									},
									Action: policy.Value{
										"sts:AssumeRole",
									},
								},
							},
						},
					},
				).
				Build(),
		},
		{
			Name:  "user_valid",
			Input: `../../../testdata/config_loading/user_valid.json`,
			Want: entities.NewBuilder().
				WithPrincipals(
					entities.Principal{
						Type:      "AWS::IAM::User",
						AccountId: "000000000000",
						Arn:       "arn:aws:iam::000000000000:user/myuser",
						Tags:      []entities.Tag{},
						InlinePolicies: []policy.Policy{
							{
								Version: "2012-10-17",
								Statement: policy.StatementBlock{
									policy.Statement{
										Sid:    "Statement1",
										Effect: "Allow",
										Action: policy.Value{
											"s3:listbucket",
										},
										Resource: policy.Value{
											"arn:aws:s3:::mybucket5",
										},
									},
								},
							},
						},
						AttachedPolicies: []entities.Arn{
							"arn:aws:iam::000000000000:policy/Shared",
						},
						Groups: []entities.Arn{
							"arn:aws:iam::000000000000:group/family",
						},
					},
				).
				WithResources(
					entities.Resource{
						Type:      "AWS::IAM::User",
						AccountId: "000000000000",
						Arn:       "arn:aws:iam::000000000000:user/myuser",
						Tags:      []entities.Tag{},
					},
				).
				Build(),
		},
		{
			Name:  "bucket_valid",
			Input: `../../../testdata/config_loading/bucket_valid.json`,
			Want: entities.NewBuilder().
				WithResources(
					entities.Resource{
						Type:      "AWS::S3::Bucket",
						AccountId: "000000000000",
						Arn:       "arn:aws:s3:::somebucket",
						Tags: []entities.Tag{
							{
								Key:   "this-bucket-tag",
								Value: "is-cool",
							},
						},
						Policy: policy.Policy{
							Version: "2012-10-17",
							Statement: policy.StatementBlock{
								policy.Statement{
									Sid:    "AllowGetObject",
									Effect: "Allow",
									Principal: policy.Principal{
										AWS: policy.Value{
											"arn:aws:iam::000000000000:role/nsiow",
										},
									},
									Action: policy.Value{
										"s3:GetObject",
									},
									Resource: policy.Value{
										"arn:aws:s3:::somebucket/*",
									},
								},
							},
						},
					},
				).
				Build(),
		},
		{
			Name:  "table_valid",
			Input: `../../../testdata/config_loading/table_valid.json`,
			Want: entities.NewBuilder().
				WithResources(
					entities.Resource{
						Type:      "AWS::DynamoDB::Table",
						AccountId: "000000000000",
						Region:    "us-east-1",
						Arn:       "arn:aws:dynamodb:us-east-1:000000000000:table/SomeTable",
						Tags:      []entities.Tag{},
					},
				).
				Build(),
		},
		{
			Name:  "topic_valid",
			Input: `../../../testdata/config_loading/topic_valid.json`,
			Want: entities.NewBuilder().
				WithResources(
					entities.Resource{
						Type:      "AWS::SNS::Topic",
						AccountId: "999999999999",
						Region:    "us-west-2",
						Arn:       "arn:aws:sns:us-west-2:999999999999:SimpleTopic",
						Tags:      []entities.Tag{},
						Policy: policy.Policy{
							Version: "2012-10-17",
							Id:      "__default_policy_ID",
							Statement: policy.StatementBlock{
								policy.Statement{
									Sid:    "__default_statement_ID",
									Effect: "Deny",
									Principal: policy.Principal{
										AWS: policy.Value{
											"*",
										},
									},
									Action: policy.Value{
										"SNS:Subscribe",
									},
									Resource: policy.Value{
										"arn:aws:sns:us-west-2:999999999999:SimpleTopic",
									},
								},
							},
						},
					},
				).
				Build(),
		},
		{
			Name:  "queue_valid",
			Input: `../../../testdata/config_loading/queue_valid.json`,
			Want: entities.NewBuilder().
				WithResources(
					entities.Resource{
						Type:      "AWS::SQS::Queue",
						AccountId: "000000000000",
						Region:    "us-west-2",
						Arn:       "arn:aws:sqs:us-west-2:000000000000:ExampleQueue",
						Tags:      []entities.Tag{},
						Policy: policy.Policy{
							Version: "2012-10-17",
							Statement: policy.StatementBlock{
								policy.Statement{
									Sid:    "AllowReceiveMessage",
									Effect: "Allow",
									Principal: policy.Principal{
										AWS: policy.Value{
											"arn:aws:iam::000000000000:role/nsiow",
										},
									},
									Action: policy.Value{
										"sqs:ReceiveMessage",
									},
									Resource: policy.Value{
										"arn:aws:sqs:us-west-2:000000000000:ExampleQueue",
									},
								},
							},
						},
					},
				).
				Build(),
		},
		{
			Name:  "key_valid",
			Input: `../../../testdata/config_loading/key_valid.json`,
			Want: entities.NewBuilder().
				WithResources(
					entities.Resource{
						Type:      "AWS::KMS::Key",
						AccountId: "999999999999",
						Region:    "us-west-2",
						Arn:       "arn:aws:kms:us-west-2:999999999999:key/1234abcd-12ab-34cd-56ef-1234567890ab",
						Tags:      []entities.Tag{},
					},
				).
				Build(),
		},

		// ---------------------------------------------------------------------------------------------
		// Invalid
		// ---------------------------------------------------------------------------------------------

		{
			Name:      "base_invalid",
			Input:     `../../../testdata/config_loading/base_invalid.json`,
			ShouldErr: true,
		},
		{
			Name:      "base_invalid_jsonl",
			Input:     `../../../testdata/config_loading/base_invalid.jsonl`,
			ShouldErr: true,
		},
		{
			Name:      "generic_resource_invalid_bad_region",
			Input:     `../../../testdata/config_loading/generic_resource_invalid_bad_region.json`,
			ShouldErr: true,
		},
		{
			Name:      "account_invalid_scp",
			Input:     `../../../testdata/config_loading/account_invalid_scp.json`,
			ShouldErr: true,
		},
		{
			Name:      "scp_invalid_syntax",
			Input:     `../../../testdata/config_loading/scp_invalid_syntax.json`,
			ShouldErr: true,
		},
		{
			Name:      "scp_invalid_syntax_2",
			Input:     `../../../testdata/config_loading/scp_invalid_syntax_2.json`,
			ShouldErr: true,
		},
		{
			Name:      "group_invalid_bad_shape",
			Input:     `../../../testdata/config_loading/group_invalid_bad_shape.json`,
			ShouldErr: true,
		},
		{
			Name:      "policy_invalid_json",
			Input:     `../../../testdata/config_loading/policy_invalid_json.json`,
			ShouldErr: true,
		},
		{
			Name:      "policy_invalid_no_default_version",
			Input:     `../../../testdata/config_loading/policy_invalid_no_default_version.json`,
			ShouldErr: true,
		},
		{
			Name:      "role_invalid_bad_policy",
			Input:     `../../../testdata/config_loading/role_invalid_bad_policy.json`,
			ShouldErr: true,
		},
		{
			Name:      "user_invalid_bad_inline_policy",
			Input:     `../../../testdata/config_loading/user_invalid_bad_inline_policy.json`,
			ShouldErr: true,
		},
		{
			Name:      "bucket_invalid_bad_policy",
			Input:     `../../../testdata/config_loading/bucket_invalid_bad_policy.json`,
			ShouldErr: true,
		},
		{
			Name:      "table_invalid_bad_region",
			Input:     `../../../testdata/config_loading/table_invalid_bad_region.json`,
			ShouldErr: true,
		},
		{
			Name:      "topic_invalid_bad_policy",
			Input:     `../../../testdata/config_loading/topic_invalid_bad_policy.json`,
			ShouldErr: true,
		},
		{
			Name:      "queue_invalid_bad_policy",
			Input:     `../../../testdata/config_loading/queue_invalid_bad_policy.json`,
			ShouldErr: true,
		},
		{
			Name:      "key_invalid_bad_region",
			Input:     `../../../testdata/config_loading/key_invalid_bad_region.json`,
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(fp string) (*entities.Universe, error) {
		// Load test data
		f, err := os.Open(fp)
		if err != nil {
			t.Fatalf("error while attempting to open test file '%s': %v", fp, err)
		}

		// Call the correct loader based on input type
		l := NewLoader()
		ext := path.Ext(fp)
		switch ext {
		case ".json":
			err = l.LoadJson(f)
		case ".jsonl":
			err = l.LoadJsonl(f)
		default:
			t.Fatalf("unsure how to handle ext '%s'", ext)
		}

		// Handle loading errors; these may be expected
		if err != nil {
			return nil, err
		}
		return l.Universe(), nil
	})
}

func TestLoad_EdgeCases(t *testing.T) {
	reader := &testlib.FailReader{}
	l := NewLoader()

	err := l.LoadJson(reader)
	if err == nil {
		t.Fatalf("LoadJson should have failed, but succeeded")
	}

	err = l.LoadJsonl(reader)
	if err == nil {
		t.Fatalf("LoadJson; should have failed, but succeeded")
	}
}
