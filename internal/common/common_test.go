package common

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestMap_String(t *testing.T) {
	type input struct {
		in       []string
		function func(string) string
	}

	tests := []testlib.TestCase[input, []string]{
		{
			Input: input{
				in: []string{
					"foo",
					"bar",
					"baz",
				},
				function: func(s string) string {
					return s + "0"
				},
			},
			Want: []string{
				"foo0",
				"bar0",
				"baz0",
			},
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) ([]string, error) {
		return Map(i.in, i.function), nil
	})
}
