package entities

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestService(t *testing.T) {
	tests := []testlib.TestCase[Resource, string]{
		{
			Name:  "valid_bucket",
			Input: Resource{Type: "AWS::S3::Bucket"},
			Want:  "s3",
		},
		{
			Name:  "valid_table",
			Input: Resource{Type: "AWS::DynamoDB::Table"},
			Want:  "dynamodb",
		},
		{
			Name:      "invalid_too_many_parts",
			Input:     Resource{Type: "AWS::S3::Bucket::Bad"},
			ShouldErr: true,
		},
		{
			Name:      "invalid_too_few_parts",
			Input:     Resource{Type: "AWS::S3"},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(r Resource) (string, error) {
		return r.Service()
	})
}

func TestSubresourceArn(t *testing.T) {
	type input struct {
		resource Resource
		subpath  string
	}

	tests := []testlib.TestCase[input, string]{
		{
			Name: "valid_object_no_leading_no_trailing",
			Input: input{
				resource: Resource{Arn: "arn:aws:s3:::bucket1"},
				subpath:  "foo",
			},
			Want: "arn:aws:s3:::bucket1/foo",
		},
		{
			Name: "valid_object_leading",
			Input: input{
				resource: Resource{Arn: "arn:aws:s3:::bucket1"},
				subpath:  "/foo",
			},
			Want: "arn:aws:s3:::bucket1/foo",
		},
		{
			Name: "valid_object_trailing",
			Input: input{
				resource: Resource{Arn: "arn:aws:s3:::bucket1/"},
				subpath:  "foo",
			},
			Want: "arn:aws:s3:::bucket1/foo",
		},
		{
			Name: "valid_object_leading_trailing",
			Input: input{
				resource: Resource{Arn: "arn:aws:s3:::bucket1/"},
				subpath:  "/foo",
			},
			Want: "arn:aws:s3:::bucket1/foo",
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (string, error) {
		return i.resource.SubresourceArn(i.subpath), nil
	})
}
