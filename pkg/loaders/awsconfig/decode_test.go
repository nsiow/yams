package awsconfig

import (
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
	"github.com/nsiow/yams/pkg/policy"
)

// TestDecodePolicyString confirms correct decoding of both valid and invalid policy strings
func TestDecodePolicyString(t *testing.T) {
	tests := []testrunner.TestCase[string, policy.Policy]{
		{
			Name:  "null",
			Input: `null`,
			Want:  policy.Policy{},
		},
		{
			Name:  "null_quoted",
			Input: `"null"`,
			Want:  policy.Policy{},
		},
		{
			Name: "s3read",
			Input: `
				{
				  "Version": "2012-10-17",
				  "Id": "s3read",
				  "Statement": [
				    {
				      "Effect": "Allow",
				      "Action": [
				        "s3:GetObject",
				        "s3:ListBucket"
				      ],
				      "Resource": [
				        "arn:aws:s3:::foo-bucket",
				        "arn:aws:s3:::foo-bucket/*"
				      ]
				    }
				  ]
				}
			`,
			Want: policy.Policy{
				Version: "2012-10-17",
				Id:      "s3read",
				Statement: []policy.Statement{
					{
						Effect: "Allow",
						Action: []string{
							"s3:GetObject",
							"s3:ListBucket",
						},
						Resource: []string{
							"arn:aws:s3:::foo-bucket",
							"arn:aws:s3:::foo-bucket/*",
						},
					},
				},
			},
		},
		{
			Name:  "s3read_escaped",
			Input: `%7B%22Version%22%3A%222012-10-17%22%2C%22Id%22%3A%22s3read%22%2C%22Statement%22%3A%5B%7B%22Effect%22%3A%22Allow%22%2C%22Action%22%3A%5B%22s3%3AGetObject%22%2C%22s3%3AListBucket%22%5D%2C%22Resource%22%3A%5B%22arn%3Aaws%3As3%3A%3A%3Afoo-bucket%22%2C%22arn%3Aaws%3As3%3A%3A%3Afoo-bucket%2F%2A%22%5D%7D%5D%7D`,
			Want: policy.Policy{
				Version: "2012-10-17",
				Id:      "s3read",
				Statement: []policy.Statement{
					{
						Effect: "Allow",
						Action: []string{
							"s3:GetObject",
							"s3:ListBucket",
						},
						Resource: []string{
							"arn:aws:s3:::foo-bucket",
							"arn:aws:s3:::foo-bucket/*",
						},
					},
				},
			},
		},
		{
			Name:  "s3read_escaped_quoted",
			Input: `"%7B%22Version%22%3A%222012-10-17%22%2C%22Id%22%3A%22s3read%22%2C%22Statement%22%3A%5B%7B%22Effect%22%3A%22Allow%22%2C%22Action%22%3A%5B%22s3%3AGetObject%22%2C%22s3%3AListBucket%22%5D%2C%22Resource%22%3A%5B%22arn%3Aaws%3As3%3A%3A%3Afoo-bucket%22%2C%22arn%3Aaws%3As3%3A%3A%3Afoo-bucket%2F%2A%22%5D%7D%5D%7D"`,
			Want: policy.Policy{
				Version: "2012-10-17",
				Id:      "s3read",
				Statement: []policy.Statement{
					{
						Effect: "Allow",
						Action: []string{
							"s3:GetObject",
							"s3:ListBucket",
						},
						Resource: []string{
							"arn:aws:s3:::foo-bucket",
							"arn:aws:s3:::foo-bucket/*",
						},
					},
				},
			},
		},
		{
			Name:      "invalid_empty_quoted",
			Input:     `""`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_escaping",
			Input:     `%+10`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_policy_1",
			Input:     `["This", "is", "not", "a", "policy"]`,
			ShouldErr: true,
		},
		{
			Name: "invalid_policy_2",
			Input: `
				{
				  "Version": "2012-10-17",
				  "Id": "s3read",
				  "Statement": [
				    {
				      "Action": [
				        "s3:GetObject",
				      ],
				      "NotAction": [
				        "s3:ListBucket"
				      ]
				    }
				  ]
				}
			`,
			ShouldErr: true,
		},
	}

	testrunner.RunTestSuite(t, tests, decodePolicyString)
}
