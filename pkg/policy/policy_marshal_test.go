package policy

import (
	"encoding/json"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

// TestPolicyMarshal validates the process of rendering policies as JSON
func TestPolicyMarshal(t *testing.T) {
	tests := []testlib.TestCase[Policy, string]{
		{
			Input: Policy{
				Version: "2012-10-17",
				Id:      "s3read",
				Statement: []Statement{
					{
						Effect: "Allow",
						Principal: Principal{
							AWS: []string{"arn:aws:iam::12345:role/SomeRole"},
						},
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
			Want: `{"Version":"2012-10-17","Id":"s3read","Statement":[{"Sid":"","Effect":"Allow","Principal":{"AWS":"arn:aws:iam::12345:role/SomeRole"},"Action":["s3:GetObject","s3:ListBucket"],"Resource":["arn:aws:s3:::foo-bucket","arn:aws:s3:::foo-bucket/*"]}]}`,
		},
		{
			Input: Policy{
				Version: "2012-10-17",
				Id:      "s3read",
				Statement: []Statement{
					{
						Effect:    "Allow",
						Principal: Principal{All: true},
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
			Want: `{"Version":"2012-10-17","Id":"s3read","Statement":[{"Sid":"","Effect":"Allow","Principal":"*","Action":["s3:GetObject","s3:ListBucket"],"Resource":["arn:aws:s3:::foo-bucket","arn:aws:s3:::foo-bucket/*"]}]}`,
		},
	}

	testlib.RunTestSuite(t, tests, func(i Policy) (string, error) {
		b, err := json.Marshal(i)
		return string(b), err
	})
}
