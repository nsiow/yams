package wildcard

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestWildcard(t *testing.T) {
	type input struct {
		pattern    string
		value      string
		ignoreCase bool
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Input: input{
				pattern: "foo",
				value:   "foo",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "",
				value:   "",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "foo",
				value:   "bar",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "*",
				value:   "bar",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "arn:aws:s3::bucket",
				value:   "arn:aws:s3:::bucket",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "arn:aws:sns:us-east-1:*:topic",
				value:   "arn:aws:sns:us-east-1:55555:topic",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "arn:aws:sns:us-east-1::topic",
				value:   "arn:aws:sns:us-east-1:55555:topic",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "arn:aws:sns:us-east-1:55555:*-backup",
				value:   "arn:aws:sns:us-east-1:55555:topic-backup",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "arn:aws:sns:us-east-1:55555:*-backup",
				value:   "arn:aws:sns:us-east-1:55555:topicbackup",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "arn:aws:sns:us-east-1:55555:topic-for-*",
				value:   "arn:aws:sns:us-east-1:55555:topic-for-sale",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "arn:aws:sns:us-east-1:55555:topic-for-*",
				value:   "arn:aws:sns:us-east-1:55555:topic-by-sale",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "arn:aws:sns:us-east-?:123*:*-in-the-*",
				value:   "arn:aws:sns:us-east-1:12345:right-in-the-middle",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "s3:*object",
				value:   "s3:getobject",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "s3:List*",
				value:   "s3:ListBucket",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "s3:getobject",
				value:   "s3:GetObject",
			},
			Want: false,
		},
		{
			Input: input{
				pattern:    "s3:getobject",
				value:      "s3:GetObject",
				ignoreCase: true,
			},
			Want: true,
		},
		{
			Input: input{
				pattern:    "s3:*?*)][]*([][][?",
				value:      "s3:getobject",
				ignoreCase: false,
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		// Determine the correct function to call
		var got bool
		if i.ignoreCase {
			got = MatchSegmentsIgnoreCase(i.pattern, i.value)
		} else {
			got = MatchSegments(i.pattern, i.value)
		}

		return got, nil
	})
}

func TestMatchSegmentsPreSplit(t *testing.T) {
	type input struct {
		pattern       string
		valueSegments []string
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Name: "wildcard_all",
			Input: input{
				pattern:       "*",
				valueSegments: []string{"anything"},
			},
			Want: true,
		},
		{
			Name: "empty_pattern",
			Input: input{
				pattern:       "",
				valueSegments: []string{"foo"},
			},
			Want: false,
		},
		{
			Name: "exact_match",
			Input: input{
				pattern:       "arn:aws:s3:::bucket",
				valueSegments: []string{"arn", "aws", "s3", "", "", "bucket"},
			},
			Want: true,
		},
		{
			Name: "wildcard_segment",
			Input: input{
				pattern:       "arn:aws:sns:us-east-1:*:topic",
				valueSegments: []string{"arn", "aws", "sns", "us-east-1", "55555", "topic"},
			},
			Want: true,
		},
		{
			Name: "no_match",
			Input: input{
				pattern:       "arn:aws:s3:::bucket",
				valueSegments: []string{"arn", "aws", "s3", "", "", "other"},
			},
			Want: false,
		},
		{
			Name: "pattern_more_segments",
			Input: input{
				pattern:       "arn:aws:s3:::bucket:extra",
				valueSegments: []string{"arn", "aws", "s3", "", "", "bucket"},
			},
			Want: false,
		},
		{
			Name: "value_more_segments",
			Input: input{
				pattern:       "arn:aws:s3:::bucket",
				valueSegments: []string{"arn", "aws", "s3", "", "", "bucket", "extra"},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		return MatchSegmentsPreSplit(i.pattern, i.valueSegments), nil
	})
}

func TestMatchSegmentsIgnoreCase_Extended(t *testing.T) {
	type input struct {
		pattern string
		value   string
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Name: "empty_pattern",
			Input: input{
				pattern: "",
				value:   "foo",
			},
			Want: false,
		},
		{
			Name: "wildcard_all",
			Input: input{
				pattern: "*",
				value:   "anything",
			},
			Want: true,
		},
		{
			Name: "mismatched_segment_count",
			Input: input{
				pattern: "s3:getobject:extra",
				value:   "s3:GetObject",
			},
			Want: false,
		},
		{
			Name: "multi_segment_case_insensitive",
			Input: input{
				pattern: "arn:aws:s3:::BUCKET",
				value:   "arn:aws:s3:::bucket",
			},
			Want: true,
		},
		{
			Name: "suffix_wildcard_case_insensitive",
			Input: input{
				pattern: "*Object",
				value:   "getobject",
			},
			Want: true,
		},
		{
			Name: "prefix_wildcard_case_insensitive",
			Input: input{
				pattern: "S3:Get*",
				value:   "s3:getobject",
			},
			Want: true,
		},
		{
			Name: "contains_wildcard_case_insensitive",
			Input: input{
				pattern: "*Get*",
				value:   "getobject",
			},
			Want: true,
		},
		{
			Name: "mixed_wildcards_case_insensitive",
			Input: input{
				pattern: "s3:*?t*",
				value:   "s3:GetObject",
			},
			Want: true,
		},
		{
			Name: "multi_wildcard_fallthrough_case_insensitive",
			Input: input{
				pattern: "s3:G*Ob*",
				value:   "s3:getobject",
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		return MatchSegmentsIgnoreCase(i.pattern, i.value), nil
	})
}

func TestMatchViaRegex_CacheHit(t *testing.T) {
	// Call twice with same pattern to exercise the cache hit branch
	pattern := "test-cache-*-pattern"
	value := "test-cache-hit-pattern"

	result1 := MatchString(pattern, value)
	result2 := MatchString(pattern, value)

	if result1 != result2 {
		t.Fatalf("cache hit produced different result: %v vs %v", result1, result2)
	}
	if !result1 {
		t.Fatal("expected match")
	}
}

func TestMatchAllOrNothing(t *testing.T) {
	type input struct {
		pattern string
		value   string
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Input: input{
				pattern: "*",
				value:   "anything",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "arn:aws:iam::88888:role/somerole",
				value:   "arn:aws:iam::88888:role/somerole",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "arn:aws:iam::88888:role/somerole",
				value:   "arn:aws:iam::88888:role/*role",
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		return MatchAllOrNothing(i.pattern, i.value), nil
	})
}

func TestMatchArn(t *testing.T) {
	type input struct {
		pattern string
		value   string
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Input: input{
				pattern: "arn:aws:s3:::somebucket",
				value:   "arn:aws:s3::somebucket",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "*",
				value:   "arn:aws:sqs:us-east-1:88888:somequeue",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "*:aws:sqs:us-east-1:88888:somequeue",
				value:   "arn:aws:sqs:us-east-1:88888:somequeue",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "arn:*:sqs:us-east-1:88888:somequeue",
				value:   "arn:aws:sqs:us-east-1:88888:somequeue",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "arn:aws:*:us-east-1:88888:somequeue",
				value:   "arn:aws:sqs:us-east-1:88888:somequeue",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "arn:aws:sqs:*:88888:somequeue",
				value:   "arn:aws:sqs:us-east-1:88888:somequeue",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "arn:aws:sqs:us-west-*:88888:somequeue",
				value:   "arn:aws:sqs:us-east-1:88888:somequeue",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "arn:aws:sqs:us-east-1:*:somequeue",
				value:   "arn:aws:sqs:us-east-1:88888:somequeue",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "arn:aws:sqs:us-east-1:88888:*",
				value:   "arn:aws:sqs:us-east-1:88888:somequeue",
			},
			Want: true,
		},
		{
			Input: input{
				pattern: "arn:aws:iam::88888:role/somerole",
				value:   "arn:aws:iam::88888:user/somerole",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "arn:aws:iam::88888:role/somerole",
				value:   "arn:aws:iam::88888:somerole",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "arn:aws:iam::88888:role/somerole",
				value:   "arn:aws:iam::88888:role/otherrole",
			},
			Want: false,
		},
		{
			Input: input{
				pattern: "arn:aws:iam::88888:*/somerole",
				value:   "arn:aws:iam::88888:role/somerole",
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		return MatchArn(i.pattern, i.value), nil
	})
}
