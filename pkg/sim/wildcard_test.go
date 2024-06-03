package sim

import (
	"testing"
)

func TestWildcard(t *testing.T) {
	type test struct {
		pattern    string
		value      string
		want       bool
		ignoreCase bool
	}

	tests := []test{
		{
			pattern: "foo",
			value:   "foo",
			want:    true,
		},
		{
			pattern: "foo",
			value:   "bar",
			want:    false,
		},
		{
			pattern: "*",
			value:   "bar",
			want:    true,
		},
		{
			pattern: "arn:aws:s3::bucket",
			value:   "arn:aws:s3:::bucket",
			want:    false,
		},
		{
			pattern: "arn:aws:sns:us-east-1:*:topic",
			value:   "arn:aws:sns:us-east-1:55555:topic",
			want:    true,
		},
		{
			pattern: "arn:aws:sns:us-east-1::topic",
			value:   "arn:aws:sns:us-east-1:55555:topic",
			want:    false,
		},
		{
			pattern: "arn:aws:sns:us-east-1:55555:*-backup",
			value:   "arn:aws:sns:us-east-1:55555:topic-backup",
			want:    true,
		},
		{
			pattern: "arn:aws:sns:us-east-1:55555:*-backup",
			value:   "arn:aws:sns:us-east-1:55555:topicbackup",
			want:    false,
		},
		{
			pattern: "arn:aws:sns:us-east-1:55555:topic-for-*",
			value:   "arn:aws:sns:us-east-1:55555:topic-for-sale",
			want:    true,
		},
		{
			pattern: "arn:aws:sns:us-east-1:55555:topic-for-*",
			value:   "arn:aws:sns:us-east-1:55555:topic-by-sale",
			want:    false,
		},
		{
			pattern: "arn:aws:sns:us-east-?:123*:*-in-the-*",
			value:   "arn:aws:sns:us-east-1:12345:right-in-the-middle",
			want:    true,
		},
		{
			pattern: "s3:*object",
			value:   "s3:getobject",
			want:    true,
		},
		{
			pattern: "s3:List*",
			value:   "s3:ListBucket",
			want:    true,
		},
		{
			pattern: "s3:getobject",
			value:   "s3:GetObject",
			want:    false,
		},
		{
			pattern:    "s3:getobject",
			value:      "s3:GetObject",
			ignoreCase: true,
			want:       true,
		},
		{
			pattern:    "s3:*?*)][]*([][][?",
			value:      "s3:getobject",
			ignoreCase: false,
		},
	}

	for _, tc := range tests {
		// Check results
		var got bool
		if tc.ignoreCase {
			got = matchWildcardIgnoreCase(tc.pattern, tc.value)
		} else {
			got = matchWildcard(tc.pattern, tc.value)
		}

		if got != tc.want {
			t.Fatalf("failed test case, wanted %v got %v for test case: %+v", tc.want, got, tc)
		}
	}
}
