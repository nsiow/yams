package types

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestShortName(t *testing.T) {
	tests := []testlib.TestCase[Action, string]{
		{
			Input: Action{
				Service: "s3",
				Name:    "getobject",
			},
			Want: "s3:getobject",
		},
		{
			Input: Action{
				Service: "sqs",
				Name:    "LISTQUEUES",
			},
			Want: "sqs:LISTQUEUES",
		},
	}

	testlib.RunTestSuite(t, tests, func(i Action) (string, error) {
		return i.ShortName(), nil
	})
}

func TestHasTargets(t *testing.T) {
	tests := []testlib.TestCase[Action, bool]{
		{
			Name: "no_resources",
			Input: Action{
				Service:   "s3",
				Name:      "ListBuckets",
				Resources: nil,
			},
			Want: false,
		},
		{
			Name: "empty_resources",
			Input: Action{
				Service:   "s3",
				Name:      "ListBuckets",
				Resources: []Resource{},
			},
			Want: false,
		},
		{
			Name: "with_resources",
			Input: Action{
				Service: "s3",
				Name:    "GetObject",
				Resources: []Resource{
					{Name: "bucket", ARNFormats: []string{"arn:aws:s3:::*"}},
				},
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(a Action) (bool, error) {
		return a.HasTargets(), nil
	})
}

func TestTargets(t *testing.T) {
	action := Action{
		Service: "s3",
		Name:    "GetObject",
		Resources: []Resource{
			{
				Name:       "bucket",
				ARNFormats: []string{"arn:aws:s3:::*"},
			},
			{
				Name:       "object",
				ARNFormats: []string{"arn:aws:s3:::*/*"},
			},
		},
	}

	tests := []testlib.TestCase[string, bool]{
		{
			Name:  "matches_bucket",
			Input: "arn:aws:s3:::mybucket",
			Want:  true,
		},
		{
			Name:  "matches_object",
			Input: "arn:aws:s3:::mybucket/mykey",
			Want:  true,
		},
		{
			Name:  "no_match_wrong_service",
			Input: "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			Want:  false,
		},
		{
			Name:  "no_match_invalid_arn",
			Input: "not-an-arn",
			Want:  false,
		},
	}

	testlib.RunTestSuite(t, tests, func(arn string) (bool, error) {
		return action.Targets(arn), nil
	})
}

func TestTargets_NoResources(t *testing.T) {
	action := Action{
		Service:   "s3",
		Name:      "ListBuckets",
		Resources: nil,
	}

	if action.Targets("arn:aws:s3:::mybucket") {
		t.Fatal("expected false for action without resources")
	}
}
