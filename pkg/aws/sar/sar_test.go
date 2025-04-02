package sar

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

// TestMap validates the correct retrieval of the full SAR map
func TestMap(t *testing.T) {
	idx := sarIndex()

	// Validate that we have a minimum number of entries
	if len(idx) < 400 {
		t.Fatalf("test failed, SAR map did not have enough entries")
	}
}

// TestMapContents spotchecks a handful of SAR items
func TestMapContents(t *testing.T) {
	idx := sarIndex()

	tests := []testlib.TestCase[string, bool]{
		{
			Input: "foo",
			Want:  false,
		},
		{
			Input: "account",
			Want:  true,
		},
		{
			Input: "ec2",
			Want:  true,
		},
		{
			Input: "s3",
			Want:  true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i string) (bool, error) {
		_, exists := idx[i]
		return exists, nil
	})
}

// TestAll validates the correct retrieval of the full SAR dataset
func TestAll(t *testing.T) {
	numEntries := 0
	for _, service := range sar() {
		for range service.Actions {
			numEntries += 1
		}
	}

	// Validate that we have a minimum number of entries
	if numEntries < 15_000 {
		t.Fatalf("test failed, API call list did not have enough entries (%d)", numEntries)
	}
}

// TestQuery validates the correct retrieval of API calls via filtering
func TestQuery(t *testing.T) {
	tests := []testlib.TestCase[func(*Query) *Query, []string]{
		{
			Input: func(q *Query) *Query {
				return q
			},
			Want: []string{},
		},
		{
			Input: func(q *Query) *Query {
				return q.WithService("s3").WithName("getobject")
			},
			Want: []string{"GetObject"},
		},
		{
			Input: func(q *Query) *Query {
				return q.WithService("S3").WithName("getObject")
			},
			Want: []string{"GetObject"},
		},
		{
			Input: func(q *Query) *Query {
				return q.WithService("aCcOuNt")
			},
			Want: []string{"AcceptPrimaryEmailUpdate", "CloseAccount", "DeleteAlternateContact", "DisableRegion", "EnableRegion", "GetAccountInformation", "GetAlternateContact", "GetContactInformation", "GetPrimaryEmail", "GetRegionOptStatus", "ListRegions", "PutAlternateContact", "PutContactInformation", "StartPrimaryEmailUpdate"},
		},
		{
			Input: func(q *Query) *Query {
				return q.WithService("account")
			},
			Want: []string{"AcceptPrimaryEmailUpdate", "CloseAccount", "DeleteAlternateContact", "DisableRegion", "EnableRegion", "GetAccountInformation", "GetAlternateContact", "GetContactInformation", "GetPrimaryEmail", "GetRegionOptStatus", "ListRegions", "PutAlternateContact", "PutContactInformation", "StartPrimaryEmailUpdate"},
		},
	}

	testlib.RunTestSuite(t, tests, func(f func(*Query) *Query) ([]string, error) {
		q := f(NewQuery())
		t.Logf("created query: %s", q.String())

		actionNames := []string{}
		for _, action := range q.Results() {
			actionNames = append(actionNames, action.Name)
		}

		return actionNames, nil
	})
}
