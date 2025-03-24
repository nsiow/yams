package sar

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

// TestMap validates the correct retrieval of the full SAR map
func TestMap(t *testing.T) {
	sarmap := Map()

	// Validate that we have a minimum number of entries
	if len(sarmap) < 400 {
		t.Fatalf("test failed, SAR map did not have enough entries")
	}
}

// TestMapContents spotchecks a handful of SAR items
func TestMapContents(t *testing.T) {
	sarmap := Map()

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
		_, exists := sarmap[i]
		return exists, nil
	})
}

// TestAll validates the correct retrieval of the full managed policy set
func TestAll(t *testing.T) {
	apicalls := All()

	numEntries := 0
	for range apicalls {
		numEntries += 1
	}

	// Validate that we have a minimum number of entries
	if numEntries < 25_000 {
		t.Fatalf("test failed, API call list did not have enough entries (%d)", numEntries)
	}
}

// TestIter exercises the iteration behavior of the API call iteration
func TestIter(t *testing.T) {
	apicalls := All()

	numEntries := 0
	for range apicalls {
		numEntries += 1
		if numEntries == 50 {
			break
		}
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
				return q.WithService("aCcOuNt").WithAccessLevel(ACCESS_LEVEL_WRITE)
			},
			Want: []string{"AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "CloseAccount", "DeleteAlternateContact", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "DisableRegion", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "EnableRegion", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "PutAlternateContact", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "PutContactInformation", "AcceptPrimaryEmailUpdate", "StartPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate"},
		},
		{
			Input: func(q *Query) *Query {
				return q.WithService("aCcOuNt")
			},
			Want: []string{"AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "CloseAccount", "DeleteAlternateContact", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "DisableRegion", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "EnableRegion", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "GetAccountInformation", "GetAlternateContact", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "GetContactInformation", "AcceptPrimaryEmailUpdate", "GetPrimaryEmail", "GetRegionOptStatus", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "ListRegions", "AcceptPrimaryEmailUpdate", "PutAlternateContact", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "PutContactInformation", "AcceptPrimaryEmailUpdate", "StartPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate"},
		},
		{
			Input: func(q *Query) *Query {
				return q.WithService("account")
			},
			Want: []string{"AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "CloseAccount", "DeleteAlternateContact", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "DisableRegion", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "EnableRegion", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "GetAccountInformation", "GetAlternateContact", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "GetContactInformation", "AcceptPrimaryEmailUpdate", "GetPrimaryEmail", "GetRegionOptStatus", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "ListRegions", "AcceptPrimaryEmailUpdate", "PutAlternateContact", "AcceptPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate", "PutContactInformation", "AcceptPrimaryEmailUpdate", "StartPrimaryEmailUpdate", "AcceptPrimaryEmailUpdate"},
		},
	}

	testlib.RunTestSuite(t, tests, func(f func(*Query) *Query) ([]string, error) {
		q := f(NewQuery())
		t.Logf("created query: %s", q.String())

		actionNames := []string{}
		for _, call := range q.Results() {
			actionNames = append(actionNames, call.Action)
		}

		return actionNames, nil
	})
}
