package sar

import (
	"strings"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestLookup(t *testing.T) {
	type input struct {
		service string
		name    string
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Input: input{
				service: "s3",
				name:    "getobject",
			},
			Want: true,
		},
		{
			Input: input{
				service: "S3",
				name:    "GETOBJECT",
			},
			Want: true,
		},
		{
			Input: input{
				service: "sqs",
				name:    "lIsTquEuEs",
			},
			Want: true,
		},
		{
			Input: input{
				service: "foo",
				name:    "bar",
			},
			Want: false,
		},
		{
			Input: input{
				service: "s3",
				name:    "listqueues",
			},
			Want: false,
		},
	}

	// Lookup
	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		action, exists := Lookup(i.service, i.name)
		// if it doesn't exist, perform no validation
		if !exists {
			return false, nil
		}

		// if it does exist, ensure that the values equal provided parameters
		return strings.EqualFold(action.Service, i.service) &&
			strings.EqualFold(action.Name, i.name), nil
	})

	// LookupString
	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		action, exists := LookupString(i.service + ":" + i.name)
		// if it doesn't exist, perform no validation
		if !exists {
			return false, nil
		}

		// if it does exist, ensure that the values equal provided parameters
		return strings.EqualFold(action.Service, i.service) &&
			strings.EqualFold(action.Name, i.name), nil
	})
}

func TestLookupStringInvalid(t *testing.T) {
	tests := []testlib.TestCase[string, bool]{
		{
			Input: "foo:bar:baz",
			Want:  false,
		},
		{
			Input: "s3_listbuckets",
			Want:  false,
		},
	}

	// LookupString
	testlib.RunTestSuite(t, tests, func(i string) (bool, error) {
		_, exists := LookupString(i)
		return exists, nil
	})
}

func TestMustLookupString(t *testing.T) {
	tests := []testlib.TestCase[string, bool]{
		{
			Input: "s3:getobject",
			Want:  true,
		},
		{
			Input: "S3:GETOBJECT",
			Want:  true,
		},
		{
			Input: "sqs:lIsTquEuEs",
			Want:  true,
		},
	}

	// MustLookupString
	testlib.RunTestSuite(t, tests, func(i string) (bool, error) {
		action := MustLookupString(i)
		return strings.EqualFold(i, action.ShortName()), nil
	})
}

func TestMustLookupStringInvalid(t *testing.T) {
	tests := []testlib.TestCase[string, any]{
		{
			Input: "foo:bar:baz",
		},
		{
			Input: "s3_listbuckets",
		},
	}

	// LookupString
	testlib.RunTestSuite(t, tests, func(i string) (any, error) {
		defer testlib.AssertPanicWithText(t, "unable to resolve service:action from SAR: '.*'")
		action := MustLookupString(i)
		return action, nil
	})
}

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
