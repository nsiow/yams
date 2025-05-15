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
