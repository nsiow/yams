package managedpolicies

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

// TestMap validates the correct retrieval of the full managed policy map
func TestMap(t *testing.T) {
	pmap := Map()

	// Validate that we have a minimum number of entries
	if len(pmap) < 1300 {
		t.Fatalf("test failed, managed policy map did not have enough entries")
	}
}

// TestMapContents spotchecks a handful of managed policy map items
func TestMapContents(t *testing.T) {
	pmap := Map()

	tests := []testlib.TestCase[string, bool]{
		{
			Input: "foo",
			Want:  false,
		},
		{
			Input: "arn:aws:iam::aws:policy/SomeRandomAccess",
			Want:  false,
		},
		{
			Input: "arn:aws:iam::aws:policy/ReadOnlyAccess",
			Want:  true,
		},
		{
			Input: "arn:aws:iam::aws:policy/AmazonS3FullAccess",
			Want:  true,
		},
		{
			Input: "arn:aws:iam::aws:policy/ServiceQuotasReadOnlyAccess",
			Want:  true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i string) (bool, error) {
		_, exists := pmap[i]
		return exists, nil
	})
}

// TestAll validates the correct retrieval of the full managed policy set
func TestAll(t *testing.T) {
	plist := All()

	// Validate that we have a minimum number of entries
	if len(plist) < 1300 {
		t.Fatalf("test failed, managed policy list did not have enough entries")
	}
}

// TestGet validates the correct retrieval of specific managed policies
func TestGet(t *testing.T) {
	tests := []testlib.TestCase[string, bool]{
		{
			Input: "foo",
			Want:  false,
		},
		{
			Input: "arn:aws:iam::aws:policy/SomeRandomAccess",
			Want:  false,
		},
		{
			Input: "arn:aws:iam::aws:policy/ReadOnlyAccess",
			Want:  true,
		},
		{
			Input: "arn:aws:iam::aws:policy/AmazonS3FullAccess",
			Want:  true,
		},
		{
			Input: "arn:aws:iam::aws:policy/ServiceQuotasReadOnlyAccess",
			Want:  true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i string) (bool, error) {
		policy, exists := Get(i)
		if !exists {
			return false, nil
		}

		return len(policy.Statement) > 0, nil
	})
}
