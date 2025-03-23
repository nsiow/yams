package awsconfig

import (
	"os"
	"path"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// TestLoadJson confirms that we can correctly load data from JSON arrays of AWS Config data
func TestLoadJson(t *testing.T) {
	tests := []testlib.TestCase[string, entities.Environment]{

		// Valid

		{
			Name:  "valid_empty_json",
			Input: `../../../testdata/environments/valid_empty.json`,
			Want: entities.Environment{
				Principals: []entities.Principal(nil),
				Resources:  []entities.Resource(nil),
			},
		},
		{
			Name:  "valid_empty_jsonl",
			Input: `../../../testdata/environments/valid_empty.jsonl`,
			Want: entities.Environment{
				Principals: []entities.Principal(nil),
				Resources:  []entities.Resource(nil),
			},
		},
		{
			Name:  "valid_simple_1_json",
			Input: `../../../testdata/environments/valid_simple_1.json`,
			Want:  simple1Output,
		},
		{
			Name:  "valid_simple_1_jsonl",
			Input: `../../../testdata/environments/valid_simple_1.jsonl`,
			Want:  simple1Output,
		},
		{
			Name:  "valid_user_1_json",
			Input: `../../../testdata/environments/valid_user_1.json`,
			Want:  user1Output,
		},
		{
			Name:  "valid_permissions_boundaries",
			Input: `../../../testdata/environments/valid_permissions_boundaries.json`,
			Want:  permissionsBoundaryOutput,
		},
		{
			Name:  "valid_scp",
			Input: `../../../testdata/environments/valid_scp.json`,
			Want:  scpOutput,
		},

		// Invalid

		{
			Name:      "invalid_json",
			Input:     `../../../testdata/environments/invalid.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_jsonl",
			Input:     `../../../testdata/environments/invalid.jsonl`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_lots_o_junk",
			Input:     `../../../testdata/environments/invalid_lots_o_junk.jsonl`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_policy_wrong_outer_type",
			Input:     `../../../testdata/environments/invalid_policy_wrong_outer_type.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_policy_no_default_version",
			Input:     `../../../testdata/environments/invalid_policy_no_default_version.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_policy_bad_document",
			Input:     `../../../testdata/environments/invalid_policy_bad_document.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_role_bad_inline",
			Input:     `../../../testdata/environments/invalid_role_bad_inline.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_role_bad_permissions_boundary",
			Input:     `../../../testdata/environments/invalid_role_bad_permissions_boundary.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_role_bad_inline_encoding",
			Input:     `../../../testdata/environments/invalid_role_bad_inline_encoding.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_role_bad_managed",
			Input:     `../../../testdata/environments/invalid_role_bad_managed.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_role_missing_managed",
			Input:     `../../../testdata/environments/invalid_role_missing_managed.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_role_missing_permissions_boundary",
			Input:     `../../../testdata/environments/invalid_role_missing_permissions_boundary.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_user_bad_inline",
			Input:     `../../../testdata/environments/invalid_user_bad_inline.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_user_bad_inline_encoding",
			Input:     `../../../testdata/environments/invalid_user_bad_inline_encoding.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_user_bad_managed",
			Input:     `../../../testdata/environments/invalid_user_bad_managed.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_user_bad_permissions_boundary",
			Input:     `../../../testdata/environments/invalid_user_bad_permissions_boundary.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_user_missing_managed",
			Input:     `../../../testdata/environments/invalid_user_missing_managed.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_user_missing_permissions_boundary",
			Input:     `../../../testdata/environments/invalid_user_missing_permissions_boundary.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_resource_bad_policy",
			Input:     `../../../testdata/environments/invalid_resource_bad_policy.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_resource_bad_policy_type",
			Input:     `../../../testdata/environments/invalid_resource_bad_policy_type.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_user_bad_group",
			Input:     `../../../testdata/environments/invalid_user_bad_group.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_user_missing_group",
			Input:     `../../../testdata/environments/invalid_user_missing_group.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_group_bad_shape",
			Input:     `../../../testdata/environments/invalid_group_bad_shape.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_group_missing_policy",
			Input:     `../../../testdata/environments/invalid_group_missing_policy.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_scp",
			Input:     `../../../testdata/environments/invalid_scp.json`,
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(fp string) (entities.Environment, error) {
		// Load test data
		data, err := os.ReadFile(fp)
		if err != nil {
			t.Fatalf("error while attempting to read test file '%s': %v", fp, err)
		}

		// Call the correct loader based on input type
		l := NewLoader()
		ext := path.Ext(fp)
		switch ext {
		case ".json":
			err = l.LoadJson(data)
		case ".jsonl":
			err = l.LoadJsonl(data)
		default:
			t.Fatalf("unsure how to handle ext '%s'", ext)
		}

		// Handle loading errors; these may be expected
		if err != nil {
			return entities.Environment{}, err
		}
		return l.Environment(), nil
	})
}

// Define some common test variables here, which we'll use across multiple tests
var simple1Output entities.Environment = entities.Environment{
	Principals: []entities.Principal{
		{
			Type:    "AWS::IAM::Role",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:role/SimpleRole1",
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
							Effect:   "Allow",
							Action:   []string{"s3:GetObject", "s3:ListBucket"},
							Resource: []string{"arn:aws:s3:::simple-bucket", "arn:aws:s3:::simple-bucket/*"},
						},
					},
				},
			},
			AttachedPolicies: []policy.Policy{
				{
					Version: "2012-10-17",
					Statement: []policy.Statement{
						{
							Effect:   "Allow",
							Action:   []string{"sqs:ReceiveMessage"},
							Resource: []string{"arn:aws:sqs:us-east-1:0000000000000:queue-2"},
						},
					},
				},
			},
		},
	},
	Resources: []entities.Resource{
		{
			Type:    "AWS::IAM::Role",
			Account: "000000000000",
			Region:  "",
			Arn:     "arn:aws:iam::000000000000:role/SimpleRole1",
			Policy: policy.Policy{
				Version: "2012-10-17",
				Statement: []policy.Statement{
					{
						Principal: policy.Principal{
							Service: policy.Value{"ec2.amazonaws.com"},
						},
						Effect: "Allow",
						Action: []string{"sts:AssumeRole"},
					},
				},
			},
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
		},
		{
			Type:    "AWS::IAM::Policy",
			Account: "000000000000",
			Region:  "",
			Arn:     "arn:aws:iam::000000000000:policy/Common",
			Policy:  policy.Policy{},
			Tags:    []entities.Tag{},
		},
		{
			Type:    "AWS::DynamoDB::Table",
			Account: "000000000000",
			Region:  "",
			Arn:     "arn:aws:dynamodb:us-east-1:000000000000:table/SomeTable",
			Policy:  policy.Policy{},
			Tags:    []entities.Tag{},
		},
		{
			Type:    "AWS::S3::Bucket",
			Account: "000000000000",
			Region:  "",
			Arn:     "arn:aws:s3:::somebucket",
			Policy: policy.Policy{
				Version: "2012-10-17",
				Statement: []policy.Statement{
					{
						Sid: "AllowGetObject",
						Principal: policy.Principal{
							AWS: policy.Value{"arn:aws:iam::000000000000:role/nsiow"},
						},
						Effect:   "Allow",
						Action:   []string{"s3:GetObject"},
						Resource: []string{"arn:aws:s3:::somebucket/*"},
					},
				},
			},
			Tags: []entities.Tag{
				{
					Key:   "this-bucket-tag",
					Value: "is-cool",
				},
			},
		},
		{
			Type:    "AWS::SQS::Queue",
			Account: "000000000000",
			Region:  "",
			Arn:     "arn:aws:sqs:us-west-2:000000000000:ExampleQueue",
			Policy: policy.Policy{
				Version: "2012-10-17",
				Statement: []policy.Statement{
					{
						Sid: "AllowReceiveMessage",
						Principal: policy.Principal{
							AWS: policy.Value{"arn:aws:iam::000000000000:role/nsiow"},
						},
						Effect:   "Allow",
						Action:   []string{"sqs:ReceiveMessage"},
						Resource: []string{"arn:aws:sqs:us-west-2:000000000000:ExampleQueue"},
					},
				},
			},
			Tags: []entities.Tag{},
		},
		{
			Type:    "AWS::SQS::Queue",
			Account: "000000000000",
			Region:  "",
			Arn:     "arn:aws:sqs:us-west-2:000000000000:SimpleQueue",
			Policy:  policy.Policy{},
			Tags:    []entities.Tag{},
		},
		{
			Type:    "AWS::SNS::Topic",
			Account: "999999999999",
			Region:  "us-west-2",
			Arn:     "arn:aws:sns:us-west-2:999999999999:SimpleTopic",
			Policy: policy.Policy{
				Version: "2012-10-17",
				Id:      "__default_policy_ID",
				Statement: []policy.Statement{
					{
						Sid:    "__default_statement_ID",
						Effect: "Deny",
						Principal: policy.Principal{
							AWS: []string{"*"},
						},
						Action:   []string{"SNS:Subscribe"},
						Resource: []string{"arn:aws:sns:us-west-2:999999999999:SimpleTopic"},
					},
				},
			},
			Tags: []entities.Tag{},
		},
	},
}

var user1Output entities.Environment = entities.Environment{
	Principals: []entities.Principal{
		{
			Type:    "AWS::IAM::User",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:user/myuser",
			Tags:    []entities.Tag{},
			InlinePolicies: []policy.Policy{
				{
					Version: "2012-10-17",
					Statement: []policy.Statement{
						{
							Sid:      "Statement1",
							Effect:   "Allow",
							Action:   []string{"s3:listbucket"},
							Resource: []string{"arn:aws:s3:::mybucket5"},
						},
					},
				},
			},
			AttachedPolicies: []policy.Policy{
				{
					Version: "2012-10-17",
					Statement: []policy.Statement{
						{
							Effect:   "Allow",
							Action:   []string{"sqs:ReceiveMessage"},
							Resource: []string{"arn:aws:sqs:us-east-1:0000000000000:queue-5"},
						},
					},
				},
			},
			GroupPolicies: []policy.Policy{
				{
					Version: "2012-10-17",
					Statement: []policy.Statement{
						{
							Effect:   "Allow",
							Action:   []string{"sqs:ReceiveMessage"},
							Resource: []string{"arn:aws:sqs:us-east-1:0000000000000:queue-2"},
						},
					},
				},
				{
					Version: "2012-10-17",
					Statement: []policy.Statement{
						{
							Sid:      "Statement1",
							Effect:   "Allow",
							Action:   []string{"s3:listbucket"},
							Resource: []string{"arn:aws:s3:::mybucket"},
						},
					},
				},
			},
		},
	},
	Resources: []entities.Resource{
		{
			Type:    "AWS::IAM::Policy",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:policy/Common",
			Policy:  policy.Policy{},
			Tags:    []entities.Tag{},
		},
		{
			Type:    "AWS::IAM::Policy",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:policy/Shared",
			Policy:  policy.Policy{},
			Tags:    []entities.Tag{},
		},
		{
			Type:    "AWS::IAM::Group",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:group/family",
			Policy:  policy.Policy{},
			Tags:    []entities.Tag{},
		},
		{
			Type:    "AWS::IAM::User",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:user/myuser",
			Policy:  policy.Policy{},
			Tags:    []entities.Tag{},
		},
	},
}

var permissionsBoundaryOutput entities.Environment = entities.Environment{
	Principals: []entities.Principal{
		{
			Type:    "AWS::IAM::User",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:user/myuser",
			Tags:    []entities.Tag{},
			PermissionsBoundary: policy.Policy{
				Version: "2012-10-17",
				Statement: []policy.Statement{
					{
						Sid:       "Statement1",
						Effect:    "Allow",
						NotAction: []string{"iam:*"},
						Resource:  []string{"*"},
					},
				},
			},
		},
		{
			Type:    "AWS::IAM::Role",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:role/myrole",
			Tags:    []entities.Tag{},
			PermissionsBoundary: policy.Policy{
				Version: "2012-10-17",
				Statement: []policy.Statement{
					{
						Sid:       "Statement1",
						Effect:    "Allow",
						NotAction: []string{"iam:*"},
						Resource:  []string{"*"},
					},
				},
			},
		},
	},
	Resources: []entities.Resource{
		{
			Type:    "AWS::IAM::Policy",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:policy/Common",
			Policy:  policy.Policy{},
			Tags:    []entities.Tag{},
		},
		{
			Type:    "AWS::IAM::User",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:user/myuser",
			Policy:  policy.Policy{},
			Tags:    []entities.Tag{},
		},
		{
			Type:    "AWS::IAM::Role",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:role/myrole",
			Policy:  policy.Policy{},
			Tags:    []entities.Tag{},
		},
	},
}

var scpOutput entities.Environment = entities.Environment{
	Principals: []entities.Principal{
		{
			Type:    "AWS::IAM::Role",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:role/myrole",
			Tags:    []entities.Tag{},
			SCPs: [][]policy.Policy{
				{
					policy.Policy{
						Id:      "arn:aws:organizations::aws:policy/service_control_policy/p-FullAWSAccess/FullAWSAccess",
						Version: "2012-10-17",
						Statement: []policy.Statement{
							{
								Effect:   "Allow",
								Action:   []string{"*"},
								Resource: []string{"*"},
							},
						},
					},
				},
				{
					policy.Policy{
						Id:      "arn:aws:organizations::000000000000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access",
						Version: "2012-10-17",
						Statement: []policy.Statement{
							{
								Effect:   "Allow",
								Action:   []string{"s3:*"},
								Resource: []string{"*"},
							},
						},
					},
				},
			},
		},
	},
	Resources: []entities.Resource{
		{
			Type:    "AWS::IAM::Role",
			Account: "000000000000",
			Arn:     "arn:aws:iam::000000000000:role/myrole",
			Policy:  policy.Policy{},
			Tags:    []entities.Tag{},
		},
	},
}
