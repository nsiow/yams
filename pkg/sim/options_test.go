package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestOptions(t *testing.T) {
	tests := []testlib.TestCase[[]OptionF, *Options]{
		{
			Input: []OptionF{},
			Want:  &Options{},
		},
		{
			Input: []OptionF{
				WithSkipServiceAuthorizationValidation(),
			},
			Want: &Options{SkipServiceAuthorizationValidation: true},
		},
		{
			Input: []OptionF{
				WithSkipServiceAuthorizationValidation(),
			},
			Want: &Options{
				SkipServiceAuthorizationValidation: true,
			},
		},
	}

	testlib.RunTestSuite(t, tests, func(i []OptionF) (*Options, error) {
		return NewOptions(i...), nil
	})
}
