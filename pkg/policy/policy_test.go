package policy

import (
	"encoding/json"
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
)

// TestPolicyGrammar confirms we got the right shape for our policy grammar
func TestPolicyGrammar(t *testing.T) {
	tests := []testrunner.TestCase[string, Policy]{
		{
			Name: "empty_policy",
			Input: `
				{
				  "Version": "",
				  "Id": "",
				  "Statement": []
				}
			`,
			Want: Policy{
				Version:   "",
				Id:        "",
				Statement: []Statement{},
			},
		},
		{
			Name: "empty_statement_map",
			Input: `
				{
				  "Statement": {}
				}
			`,
			Want: Policy{
				Version:   "",
				Id:        "",
				Statement: []Statement{{}},
			},
		},
		{
			Name: "null_statement",
			Input: `
				{
				  "Statement": null
				}
			`,
			Want: Policy{
				Version:   "",
				Id:        "",
				Statement: []Statement{},
			},
		},
		{
			Name: "invalid_small_statement",
			Input: `
				{
				  "Statement": 0
				}
			`,
			ShouldErr: true,
		},
		{
			Name: "invalid_statement_array",
			Input: `
				{
				  "Statement": [0]
				}
			`,
			ShouldErr: true,
		},
		{
			Name: "invalid_statement_map",
			Input: `
				{
				  "Statement": {
				    "Effect": 0
				  }
				}
			`,
			ShouldErr: true,
		},
		{
			Name:  "null_policy",
			Input: `null`,
			Want:  Policy{},
		},
		{
			Name: "weird_statement_block",
			Input: `
				{
				  "Statement": ""
				}
			`,
			ShouldErr: true,
		},
		{
			Name: "s3read_policy",
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
			Want: Policy{
				Version: "2012-10-17",
				Id:      "s3read",
				Statement: []Statement{
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
			Name: "valid_structured_principal",
			Input: `
				{
				  "Version": "2012-10-17",
				  "Id": "s3read",
				  "Statement": [
				    {
				      "Effect": "Allow",
				      "Principal": {
				        "AWS": [
				          "SomeValueHere"
				        ]
				      },
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
			Want: Policy{
				Version: "2012-10-17",
				Id:      "s3read",
				Statement: []Statement{
					{
						Effect: "Allow",
						Principal: Principal{
							AWS: []string{"SomeValueHere"},
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
		},
		{
			Name: "invalid_structured_principal",
			Input: `
				{
				  "Version": "2012-10-17",
				  "Id": "s3read",
				  "Statement": [
				    {
				      "Effect": "Allow",
				      "Principal": {
				        "AWS": [
				          0
				        ]
				      },
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
			ShouldErr: true,
		},
	}

	testrunner.RunTestSuite(t, tests, func(s string) (Policy, error) {
		p := Policy{}
		err := json.Unmarshal([]byte(s), &p)
		return p, err
	})
}

// TestValidate confirms correct validation behavior for parsed policy statements
func TestValidate(t *testing.T) {
	tests := []testrunner.TestCase[string, any]{
		{
			Name: "valid",
			Input: `
        {
			  	"Effect": "Allow",
			  	"Principal": "*",
			  	"Action": "*",
			  	"Resource": "*"
			  }
			`,
			ShouldErr: false,
		},
		{
			Name: "empty_statement",
			Input: `
			  {}
			`,
			ShouldErr: true,
		},
		{
			Name: "double_principal",
			Input: `
        {
					"Effect": "Allow",
					"Principal": "*",
					"NotPrincipal": "*",
					"Action": "*",
					"Resource": "*"
				}
			`,
			ShouldErr: true,
		},
		{
			Name: "double_action",
			Input: `
			  {
					"Effect": "Allow",
					"Principal": "*",
					"Action": "*",
					"NotAction": "*",
					"Resource": "*"
				}
			`,
			ShouldErr: true,
		},
		{
			Name: "double_resource",
			Input: `
        {
					"Effect": "Allow",
					"Principal": "*",
					"Action": "*",
					"Resource": "*",
					"NotResource": "*"
				}
			`,
			ShouldErr: true,
		},
	}

	testrunner.RunTestSuite(t, tests, func(s string) (any, error) {
		stmt := Statement{}
		err := json.Unmarshal([]byte(s), &stmt)
		if err != nil {
			t.Fatalf("unexpected error prior to Validate(...) function: %v", err)
		}

		return nil, stmt.Validate()
	})
}
