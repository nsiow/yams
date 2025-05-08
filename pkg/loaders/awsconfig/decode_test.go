package awsconfig

import (
	"encoding/json"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/policy"
)

func TestDecodePolicyString(t *testing.T) {
	tests := []testlib.TestCase[string, policy.Policy]{
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
			Name:  "empty_quoted",
			Input: `""`,
			Want:  policy.Policy{},
		},
		{
			Name:  "s3read",
			Input: `"{\"Version\":\"2012-10-17\",\"Id\":\"s3read\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":[\"s3:GetObject\",\"s3:ListBucket\"],\"Resource\":[\"arn:aws:s3:::foo-bucket\",\"arn:aws:s3:::foo-bucket/*\"]}]}"`,
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
			Name:  "s3read_double_quoted",
			Input: `"\"{\\\"Version\\\":\\\"2012-10-17\\\",\\\"Id\\\":\\\"s3read\\\",\\\"Statement\\\":[{\\\"Effect\\\":\\\"Allow\\\",\\\"Action\\\":[\\\"s3:GetObject\\\",\\\"s3:ListBucket\\\"],\\\"Resource\\\":[\\\"arn:aws:s3:::foo-bucket\\\",\\\"arn:aws:s3:::foo-bucket/*\\\"]}]}\""`,
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
			Name:      "invalid_unbalanced_quotes_1",
			Input:     `"`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_unbalanced_quotes_3",
			Input:     `"\""`,
			ShouldErr: true,
		},
		{
			Name:      "invalid_escaping",
			Input:     `"%+10"`,
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

	testlib.RunTestSuite(t, tests, func(s string) (policy.Policy, error) {
		var e EncodedPolicy
		err := json.Unmarshal([]byte(s), &e)
		if err != nil {
			return policy.Policy{}, err
		}

		return policy.Policy(e), nil
	})
}
