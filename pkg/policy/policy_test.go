package policy

import (
	"encoding/json"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestPolicyEmpty(t *testing.T) {
	tests := []testlib.TestCase[string, bool]{
		{
			Name: "empty_policy",
			Input: `
				{
					"Version": "",
					"Id": "",
					"Statement": []
				}
			`,
			Want: true,
		},
		{
			Name: "non_empty_policy",
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
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(s string) (bool, error) {
		p := Policy{}
		err := json.Unmarshal([]byte(s), &p)
		if err != nil {
			return false, err
		}

		return p.Empty(), nil
	})
}

func TestPolicyValid(t *testing.T) {
	tests := []testlib.TestCase[string, bool]{
		{
			Name: "valid_policies",
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
			Want: true,
		},
		{
			Name: "invalid_policy",
			Input: `
				{
					"Version": "",
					"Id": "",
					"Statement": [
						{
							"Effect": "Allow",
							"Principal": "*",
							"Action": "*",
							"Resource": "*",
							"NotResource": "*"
						}
					]
				}
			`,
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(s string) (bool, error) {
		p := Policy{}
		err := json.Unmarshal([]byte(s), &p)
		if err != nil {
			return false, err
		}

		err = p.Validate()
		return err == nil, err
	})
}
