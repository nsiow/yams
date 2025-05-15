package arn

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestArn(t *testing.T) {
	type output struct {
		Partition    string
		Service      string
		Region       string
		AccountId    string
		ResourceType string
		ResourceId   string
	}

	tests := []testlib.TestCase[string, output]{
		{
			Name:  "empty",
			Input: "",
			Want:  output{},
		},
		{
			Name:  "s3_bucket",
			Input: "arn:aws:s3:::my-bucket",
			Want: output{
				Partition:    "aws",
				Service:      "s3",
				Region:       "",
				AccountId:    "",
				ResourceType: "bucket",
				ResourceId:   "my-bucket",
			},
		},
		{
			Name:  "s3_object",
			Input: "arn:aws:s3:::my-bucket/folder/object.txt",
			Want: output{
				Partition:    "aws",
				Service:      "s3",
				Region:       "",
				AccountId:    "",
				ResourceType: "object",
				ResourceId:   "my-bucket/folder/object.txt",
			},
		},
		{
			Name:  "lambda",
			Input: "arn:aws:lambda:us-east-1:123456789012:function:my-function",
			Want: output{
				Partition:    "aws",
				Service:      "lambda",
				Region:       "us-east-1",
				AccountId:    "123456789012",
				ResourceType: "function",
				ResourceId:   "my-function",
			},
		},
		{
			Name:  "govcloud_instance",
			Input: "arn:aws-us-gov:ec2:us-gov-east-1:123456789012:instance/i-0123456789abcdef0",
			Want: output{
				Partition:    "aws-us-gov",
				Service:      "ec2",
				Region:       "us-gov-east-1",
				AccountId:    "123456789012",
				ResourceType: "instance",
				ResourceId:   "i-0123456789abcdef0",
			},
		},
		{
			Name:  "cn_ddb_table",
			Input: "arn:aws-cn:dynamodb:cn-north-1:123456789012:table/my-table",
			Want: output{
				Partition:    "aws-cn",
				Service:      "dynamodb",
				Region:       "cn-north-1",
				AccountId:    "123456789012",
				ResourceType: "table",
				ResourceId:   "my-table",
			},
		},
		{
			Name:  "not_enough_components",
			Input: "arn:aws:::",
			Want: output{
				Partition:    "aws",
				Service:      "",
				Region:       "",
				AccountId:    "",
				ResourceType: "",
				ResourceId:   "",
			},
		},
		{
			Name:  "imaginary_service",
			Input: "arn:aws:someservice:us-east-1:123456789012:foo/bar",
			Want: output{
				Partition:    "aws",
				Service:      "someservice",
				Region:       "us-east-1",
				AccountId:    "123456789012",
				ResourceType: "foo",
				ResourceId:   "bar",
			},
		},
	}

	testlib.RunTestSuite(t, tests, func(arn string) (output, error) {
		return output{
			Partition:    Partition(arn),
			Service:      Service(arn),
			Region:       Region(arn),
			AccountId:    Account(arn),
			ResourceType: ResourcePath(arn),
			ResourceId:   ResourceId(arn),
		}, nil
	})
}
