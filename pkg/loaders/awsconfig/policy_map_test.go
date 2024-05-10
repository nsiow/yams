package awsconfig

import (
	"reflect"
	"testing"

	pol "github.com/nsiow/yams/pkg/policy"
)

// TestStorageRetrieval confirms the ability to store/retrieve managed policies correctly
func TestStorageRetrieval(t *testing.T) {
	type test struct {
		name   string     // friendly name for the test
		store  bool       // whether or not to store the provided policy before the test
		exists bool       // whether or not we should expect the requested ARN to exist
		arn    string     // the ARN to use for storage + retrieval
		policy pol.Policy // the policy to store + validate unpon retrieval
	}

	// Define inputs
	tests := []test{
		{
			name:   "empty_policy",
			store:  true,
			exists: true,
			arn:    "arn:aws:iam::000000000000:policy/EmptyPolicy",
			policy: pol.Policy{},
		},
		{
			name:   "s3read_basic",
			store:  true,
			exists: true,
			arn:    "arn:aws:iam::000000000000:policy/AmazonFakeS3ReadOnlyAccess",
			policy: pol.Policy{
				Version: "2012-10-17",
				Id:      "s3read",
				Statement: []pol.Statement{
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
			name:   "normalize_1",
			store:  true,
			exists: true,
			arn:    "arn:aws:iam::aws:policy/NormalizationTest",
			policy: pol.Policy{Id: "something_unique"},
		},
		{
			name:   "normalize_2",
			store:  true,
			exists: true,
			arn:    "arn:aws:iam::aws:policy/aws-service-role/NormalizationTest",
			policy: pol.Policy{Id: "something_unique"},
		},
		{
			name:   "normalize_3",
			store:  true,
			exists: true,
			arn:    "arn:aws:iam::aws:policy/service-role/NormalizationTest",
			policy: pol.Policy{Id: "something_unique"},
		},
		{
			name:   "retrieve_only_normalize_1",
			store:  false,
			exists: true,
			arn:    "arn:aws:iam::aws:policy/NormalizationTest",
			policy: pol.Policy{Id: "something_unique"},
		},
		{
			name:   "retrieve_only_normalize_2",
			store:  false,
			exists: true,
			arn:    "arn:aws:iam::aws:policy/aws-service-role/NormalizationTest",
			policy: pol.Policy{Id: "something_unique"},
		},
		{
			name:   "retrieve_only_normalize_3",
			store:  false,
			exists: true,
			arn:    "arn:aws:iam::aws:policy/service-role/NormalizationTest",
			policy: pol.Policy{Id: "something_unique"},
		},
		{
			name:   "should_be_missing",
			store:  false,
			exists: false,
			arn:    "arn:aws:iam::aws:policy/NonexistentPolicy",
		},
	}

	// Add our test inputs
	m := NewPolicyMap()
	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		// Store the policy if requested
		if tc.store {
			m.Add(tc.arn, tc.policy)
		}

		// Retrieve the policy
		got, exists := m.Get(tc.arn)
		switch {
		case exists && tc.exists:
			// expected hit; got hit
		case !exists && tc.exists:
			t.Fatalf("expected policy, but got miss for test case %+v", tc)
		case exists && !tc.exists:
			t.Fatalf("expected miss, but policy exists for test case %+v", tc)
		case !exists && !tc.exists:
			// expected miss; got miss
		}

		// Compare retrieve to expected
		if !tc.exists {
			if !reflect.DeepEqual(tc.policy, got) {
				t.Fatalf("expected: %#v, got: %#v, for input: %+v", tc.policy, got, tc)
			}
		}
	}
}
