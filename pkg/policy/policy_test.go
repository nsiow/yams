package policy

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestPolicyGrammar confirms we got the right shape for our policy grammar
func TestPolicyGrammar(t *testing.T) {
	type test struct {
		name  string
		input string
		want  Policy
		err   bool
	}

	tests := []test{
		{
			name: "empty_policy",
			input: `
			  {
          "Version": "",
          "Id": "",
          "Statement": []
				}
			`,
			want: Policy{
				Version:   "",
				Id:        "",
				Statement: []Statement{},
			},
		},
		{
			name: "empty_statement_map",
			input: `
				{
				  "Statement": {}
				}
			`,
			want: Policy{
				Version:   "",
				Id:        "",
				Statement: []Statement{{}},
			},
		},
		{
			name: "invalid_small_statement",
			input: `
				{
				  "Statement": 0
				}
			`,
			err: true,
		},
		{
			name: "invalid_statement_array",
			input: `
				{
				  "Statement": [0]
				}
			`,
			err: true,
		},
		{
			name: "invalid_statement_map",
			input: `
				{
				  "Statement": {
				    "Effect": 0
				  }
				}
			`,
			err: true,
		},
		{
			name:  "null_policy",
			input: `null`,
			want:  Policy{},
		},
		{
			name: "weird_statement_block",
			input: `
			  {
					"Statement": ""
				}
			`,
			err: true,
		},
		{
			name: "s3read_policy",
			input: `
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
			want: Policy{
				Version: "2012-10-17",
				Id:      "s3read",
				Statement: []Statement{
					{
						Effect: "Allow",
						Action: Action{
							[]string{
								"s3:GetObject",
								"s3:ListBucket",
							},
						},
						Resource: Resource{
							[]string{
								"arn:aws:s3:::foo-bucket",
								"arn:aws:s3:::foo-bucket/*",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		// Unmarshal into a policy
		p := Policy{}
		err := json.Unmarshal([]byte(tc.input), &p)

		switch {
		case err == nil && tc.err:
			t.Fatalf("expected error, got success for asset '%s': %v", tc.name, err)
		case err != nil && tc.err:
			// expected error; got error
			t.Logf("test saw error: %v", err)
		case err == nil && !tc.err:
			// no error and not expecting one, continue
			break
		case err != nil && !tc.err:
			t.Fatalf("unable to create policy from asset '%s': %v", tc.name, err)
		}

		// Check against expected value
		if !reflect.DeepEqual(tc.want, p) {
			t.Fatalf("expected: %#v, got: %#v", tc.want, p)
		}
	}
}

// TestValidate confirms correct validation behavior for parsed policies
func TestValidate(t *testing.T) {
	type test struct {
		name  string
		input string
		err   bool
	}

	tests := []test{
		{
			name: "empty_policy",
			input: `
			  {
          "Version": "",
          "Id": "",
          "Statement": {}
				}
			`,
			err: true,
		},
		{
			name: "double_principal",
			input: `
			  {
          "Version": "",
          "Id": "",
          "Statement": {
						"Principal": "*",
						"NotPrincipal": "*"
					}
				}
			`,
			err: true,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		// Unmarshal into a policy
		p := Policy{}
		err := json.Unmarshal([]byte(tc.input), &p)
		if err != nil {
			t.Fatalf("unable to create policy from asset '%s': %v", tc.name, err)
		}

		// Validate statements
		for i, stmt := range p.Statement {
			err := stmt.Validate()
			switch {
			case err == nil && tc.err:
				t.Fatalf("expected error, got success for statement #%d, test case '%s': %v", i, tc.name, err)
			case err != nil && tc.err:
				// expected error; got error
				t.Logf("test saw error: %v", err)
			case err == nil && !tc.err:
				// no error and not expecting one, continue
				break
			case err != nil && !tc.err:
				t.Fatalf("expected success, got error for statement #%d, test case '%s': %v", i, tc.name, err)
			}
		}
	}
}
