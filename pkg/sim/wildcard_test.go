package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
)

func TestWildcard(t *testing.T) {
	type input struct {
		pattern    string
		value      string
		ignoreCase bool
	}

	tests := []testrunner.TestCase[input, bool]{
		{
			Input: input{
				pattern: "foo",
				value:   "foo",
			},
			Want: true,
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
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		// Determine the correct function to call
		var got bool
		if i.ignoreCase {
			got = matchWildcardIgnoreCase(i.pattern, i.value)
		} else {
			got = matchWildcard(i.pattern, i.value)
		}

		return got, nil
	})
}
