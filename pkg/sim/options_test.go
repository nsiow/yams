package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestOptions(t *testing.T) {
	tests := []testlib.TestCase[[]OptionF, Options]{
		{
			Input: []OptionF{},
			Want: Options{
				DefaultS3Key: "*",
			},
		},
		{
			Input: []OptionF{
				WithSkipServiceAuthorizationValidation(),
			},
			Want: Options{
				DefaultS3Key:                       "*",
				SkipServiceAuthorizationValidation: true,
			},
		},
		{
			Input: []OptionF{
				WithSkipServiceAuthorizationValidation(),
			},
			Want: Options{
				DefaultS3Key:                       "*",
				SkipServiceAuthorizationValidation: true,
			},
		},
		{
			Input: []OptionF{
				WithOverlay(SimpleTestUniverse_1),
			},
			Want: Options{
				DefaultS3Key: "*",
				Overlay:      SimpleTestUniverse_1,
			},
		},
		{
			Input: []OptionF{
				WithAdditionalProperties(
					map[string]string{
						"foo": "bar",
					},
				),
			},
			Want: Options{
				DefaultS3Key: "*",
				Context: NewBagFromMap(
					map[string]string{
						"foo": "bar",
					},
				),
			},
		},
		{
			Input: []OptionF{
				WithDefaultS3Key("something else"),
			},
			Want: Options{
				DefaultS3Key: "something else",
			},
		},
		{
			Input: []OptionF{
				WithEnableFuzzyMatchArn(),
			},
			Want: Options{
				DefaultS3Key:        "*",
				EnableFuzzyMatchArn: true,
			},
		},
		{
			Input: []OptionF{
				WithForceFailure(),
			},
			Want: Options{
				DefaultS3Key: "*",
				ForceFailure: true,
			},
		},
	}

	testlib.RunTestSuite(t, tests, func(i []OptionF) (Options, error) {
		return NewOptions(i...), nil
	})
}
