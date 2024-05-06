package awsconfig

import (
	"reflect"
	"testing"

	p "github.com/nsiow/yams/pkg/policy"
)

func TestDecodePolicyStringValid(t *testing.T) {
	type test struct {
		name  string
		input string
		want  p.Policy
	}

	tests := []test{
		{
			name:  "null",
			input: `null`,
			want:  p.Policy{},
		},
		{
			name:  "null_quoted",
			input: `"null"`,
			want:  p.Policy{},
		},
		{
			name: "s3read",
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
			want: p.Policy{
				Version: "2012-10-17",
				Id:      "s3read",
				Statement: []p.Statement{
					{
						Effect: "Allow",
						Action: p.Action{
							Value: []string{
								"s3:GetObject",
								"s3:ListBucket",
							},
						},
						Resource: p.Resource{
							Value: []string{
								"arn:aws:s3:::foo-bucket",
								"arn:aws:s3:::foo-bucket/*",
							},
						},
					},
				},
			},
		},
		{
			name:  "s3read_escaped",
			input: `%7B%22Version%22%3A%222012-10-17%22%2C%22Id%22%3A%22s3read%22%2C%22Statement%22%3A%5B%7B%22Effect%22%3A%22Allow%22%2C%22Action%22%3A%5B%22s3%3AGetObject%22%2C%22s3%3AListBucket%22%5D%2C%22Resource%22%3A%5B%22arn%3Aaws%3As3%3A%3A%3Afoo-bucket%22%2C%22arn%3Aaws%3As3%3A%3A%3Afoo-bucket%2F%2A%22%5D%7D%5D%7D`,
			want: p.Policy{
				Version: "2012-10-17",
				Id:      "s3read",
				Statement: []p.Statement{
					{
						Effect: "Allow",
						Action: p.Action{
							Value: []string{
								"s3:GetObject",
								"s3:ListBucket",
							},
						},
						Resource: p.Resource{
							Value: []string{
								"arn:aws:s3:::foo-bucket",
								"arn:aws:s3:::foo-bucket/*",
							},
						},
					},
				},
			},
		},
		{
			name:  "s3read_escaped_quoted",
			input: `"%7B%22Version%22%3A%222012-10-17%22%2C%22Id%22%3A%22s3read%22%2C%22Statement%22%3A%5B%7B%22Effect%22%3A%22Allow%22%2C%22Action%22%3A%5B%22s3%3AGetObject%22%2C%22s3%3AListBucket%22%5D%2C%22Resource%22%3A%5B%22arn%3Aaws%3As3%3A%3A%3Afoo-bucket%22%2C%22arn%3Aaws%3As3%3A%3A%3Afoo-bucket%2F%2A%22%5D%7D%5D%7D"`,
			want: p.Policy{
				Version: "2012-10-17",
				Id:      "s3read",
				Statement: []p.Statement{
					{
						Effect: "Allow",
						Action: p.Action{
							Value: []string{
								"s3:GetObject",
								"s3:ListBucket",
							},
						},
						Resource: p.Resource{
							Value: []string{
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
		t.Logf("running test: %s", tc.name)

		got, err := decodePolicyString(tc.input)
		if err != nil {
			t.Fatalf("error in test: '%s'\ninput=%v\nerr=%v", tc.name, tc.input, err)
		}

		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("assertion failed: '%s'\nexpected=%v\ngot=%v\ninput=%v", tc.name, tc.want, got, tc.input)
		}
	}
}

func TestDecodePolicyStringInvalid(t *testing.T) {
	type test struct {
		name  string
		input string
	}

	tests := []test{
		{
			name:  "empty_quoted",
			input: `""`,
		},
		{
			name:  "invalid_escaping",
			input: `%+10`,
		},
		{
			name:  "invalid_policy_1",
			input: `["This", "is", "not", "a", "policy"]`,
		},
		{
			name: "invalid_policy_2",
			input: `
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
		},
	}

	for _, tc := range tests {
		t.Logf("running test: %s", tc.name)

		_, err := decodePolicyString(tc.input)
		if err == nil {
			t.Fatalf("expected error, got success for test case '%s': %v", tc.name, err)
		} else {
			t.Logf("test '%s' saw expected error: %v", tc.name, err)
		}
	}
}
