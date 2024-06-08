package awsconfig

import (
	"os"
	"path"
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// TestLoadJsonValid confirms that we can correctly load data from JSON arrays of AWS Config data
func TestLoadJsonValid(t *testing.T) {
	tests := []testrunner.TestCase[string, entities.Environment]{

		// Valid

		{
			Name:  "empty_json",
			Input: `../../../testdata/environments/empty.json`,
			Want: entities.Environment{
				Principals: []entities.Principal(nil),
				Resources:  []entities.Resource(nil),
			},
		},
		{
			Name:  "empty_jsonl",
			Input: `../../../testdata/environments/empty.jsonl`,
			Want: entities.Environment{
				Principals: []entities.Principal(nil),
				Resources:  []entities.Resource(nil),
			},
		},
		{
			Name:  "simple_1_json",
			Input: `../../../testdata/environments/simple_1.json`,
			Want:  simple1Output,
		},
		{
			Name:  "simple_1_jsonl",
			Input: `../../../testdata/environments/simple_1.jsonl`,
			Want:  simple1Output,
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
			Input:     `../../../testdata/environments/lots_o_junk.jsonl`,
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
			Name:      "invalid_principal_bad_inline",
			Input:     `../../../testdata/environments/invalid_principal_bad_inline.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_principal_bad_inline_encoding",
			Input:     `../../../testdata/environments/invalid_principal_bad_inline_encoding.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_principal_bad_managed",
			Input:     `../../../testdata/environments/invalid_principal_bad_managed.json`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_principal_missing_managed",
			Input:     `../../../testdata/environments/invalid_principal_missing_managed.json`,
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
	}

	testrunner.RunTestSuite(t, tests, func(fp string) (entities.Environment, error) {
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
