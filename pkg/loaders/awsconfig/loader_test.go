package awsconfig

import (
	"encoding/json"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/environment"
	"github.com/nsiow/yams/pkg/policy"
)

// Define some common test variables here, which we'll use across multiple tests
var simple1Output environment.Universe = environment.Universe{
	Principals: []entities.Principal{
		{
			Type:    "AWS::IAM::Role",
			Account: "000000000000",
			Region:  "",
			Arn:     "arn:aws:iam::000000000000:role/SimpleRole1",
			Tags: []entities.Tag{
				{
					Key:   "some-business-tag",
					Value: "important-business-thing",
					Tag:   "some-business-tag=important-business-thing",
				},
				{
					Key:   "some-technical-tag",
					Value: "important-technical-thing",
					Tag:   "some-technical-tag=important-technical-thing",
				},
			},
			InlinePolicies: []policy.Policy{
				{
					Version: "2012-10-17",
					Statement: []policy.Statement{
						{
							Effect: "Allow",
							Action: policy.Action{
								Value: []string{"s3:GetObject", "s3:ListBucket"},
							},
							Resource: policy.Resource{
								Value: []string{"arn:aws:s3:::simple-bucket", "arn:aws:s3:::simple-bucket/*"},
							},
						},
					},
				},
			},
			ManagedPolicies: []policy.Policy{
				{
					Version: "2012-10-17",
					Statement: []policy.Statement{
						{
							Effect: "Allow",
							Action: policy.Action{
								Value: []string{"sqs:ReceiveMessage"},
							},
							Resource: policy.Resource{
								Value: []string{"arn:aws:sqs:us-east-1:0000000000000:queue-2"},
							},
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
						Action: policy.Action{
							Value: []string{"sts:AssumeRole"},
						},
					},
				},
			},
			Tags: []entities.Tag{
				{
					Key:   "some-business-tag",
					Value: "important-business-thing",
					Tag:   "some-business-tag=important-business-thing",
				},
				{
					Key:   "some-technical-tag",
					Value: "important-technical-thing",
					Tag:   "some-technical-tag=important-technical-thing",
				},
			},
		},
		{
			Type:    "AWS::IAM::Policy",
			Account: "000000000000",
			Region:  "",
			Arn:     "arn:aws:iam::000000000000:policy/Common",
			Policy:  policy.Policy{},
			Tags:    entities.Tags{},
		},
	},
}

// TestLoadJson confirms that we can correctly load data from JSON arrays of AWS Config data
func TestLoadJsonValid(t *testing.T) {
	type test struct {
		name  string
		input string
		want  environment.Universe
	}

	tests := []test{
		{
			name:  "empty",
			input: `../../../testdata/environments/empty.json`,
			want: environment.Universe{
				Principals: []entities.Principal(nil),
				Resources:  []entities.Resource(nil),
			},
		},
		{
			name:  "simple_1",
			input: `../../../testdata/environments/simple_1.json`,
			want:  simple1Output,
		},
	}

	for _, tc := range tests {
		subtests := []string{
			tc.input,
			tc.input + "l",
		}

		for _, input := range subtests {
			t.Logf("running test case: %s (file: %s)", tc.name, input)

			// Read requested input file
			inputBytes, err := os.ReadFile(input)
			if err != nil {
				t.Fatalf("unable to read file '%s' for test case: '%s': %v", input, tc.name, err)
			}

			// Call correct loader based on input type
			l := NewLoader()
			ext := path.Ext(input)
			switch ext {
			case ".json":
				err = l.LoadJson(inputBytes)
			case ".jsonl":
				err = l.LoadJsonl(inputBytes)
			default:
				t.Fatalf("unsure how to handle ext '%s' for test case: '%s'", ext, tc.name)
			}
			if err != nil {
				t.Fatalf("unexpected error for test case: '%s': %v", tc.name, err)
			}

			// Construct our universe based on what we received
			got := environment.Universe{
				Principals: l.Principals(),
				Resources:  l.Resources(),
			}

			// Compare and validate; pretty-print in JSON if something goes wrong for easier debugging
			if !reflect.DeepEqual(tc.want, got) {
				// TODO(nsiow) remove all string(...) casts; just use %s directly
				// across code base
				wantString, err := json.MarshalIndent(tc.want, "", " ")
				if err != nil {
					t.Logf("error while trying to pretty print test error; falling back")
					t.Fatalf("expected: %s, got: %#v for test case '%s'", tc.want, got, tc.name)
				}
				gotString, err := json.MarshalIndent(got, "", " ")
				if err != nil {
					t.Logf("error while trying to pretty print test error; falling back")
					t.Fatalf("expected: %s, got: %#v for test case '%s'", tc.want, got, tc.name)
				}
				t.Fatalf("*failure* on test '%s'\n\n*expected*\n%s\n\n*got*\n%s", tc.name, wantString, gotString)
			}
		}
	}
}
