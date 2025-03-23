package gate

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestGate(t *testing.T) {
	type input struct {
		numInverts int
		value      bool
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Name: "no_inverts_true",
			Input: input{
				numInverts: 0,
				value:      true,
			},
			Want: true,
		},
		{
			Name: "no_inverts_false",
			Input: input{
				numInverts: 0,
				value:      true,
			},
			Want: true,
		},
		{
			Name: "single_invert_true",
			Input: input{
				numInverts: 1,
				value:      true,
			},
			Want: false,
		},
		{
			Name: "single_invert_false",
			Input: input{
				numInverts: 1,
				value:      false,
			},
			Want: true,
		},
		{
			Name: "double_invert_true",
			Input: input{
				numInverts: 2,
				value:      true,
			},
			Want: true,
		},
		{
			Name: "double_invert_false",
			Input: input{
				numInverts: 2,
				value:      false,
			},
			Want: false,
		},
		{
			Name: "high_invert_even_true",
			Input: input{
				numInverts: 100,
				value:      true,
			},
			Want: true,
		},
		{
			Name: "high_invert_even_false",
			Input: input{
				numInverts: 100,
				value:      false,
			},
			Want: false,
		},
		{
			Name: "high_invert_odd_true",
			Input: input{
				numInverts: 151,
				value:      true,
			},
			Want: false,
		},
		{
			Name: "high_invert_odd_false",
			Input: input{
				numInverts: 151,
				value:      false,
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		// Invert as many times as requested
		g := Gate{}
		for j := 0; j < i.numInverts; j++ {
			g.Invert()
		}

		return g.Apply(i.value), nil
	})
}
