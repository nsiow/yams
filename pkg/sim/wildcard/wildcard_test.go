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

// TestMatchAllOrNothing validates correct matching behavior for All/Nothing matches
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

// TestMatchArn validates correct matching behavior for Arn matches
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
