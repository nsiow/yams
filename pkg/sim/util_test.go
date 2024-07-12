package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
)

// TestId validates correct resolution behavior of Id(...)
func TestId(t *testing.T) {
	type input struct {
		id  string
		idx int
	}

	tests := []testrunner.TestCase[input, string]{
		{
			Name: "empty_id",
			Input: input{
				id:  "",
				idx: 123,
			},
			Want: "123",
		},
		{
			Name: "valid_id",
			Input: input{
				id:  "foo",
				idx: 123,
			},
			Want: "foo",
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (string, error) {
		return Id(i.id, i.idx), nil
	})
}
