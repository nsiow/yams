package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/policy"
)

func TestEffectSet(t *testing.T) {
	type output struct {
		Allowed          bool
		Denied           bool
		ExplicitlyDenied bool
	}

	tests := []testlib.TestCase[[]policy.Effect, output]{
		{
			Name:  "implicit_deny",
			Input: []policy.Effect{},
			Want: output{
				Allowed:          false,
				Denied:           true,
				ExplicitlyDenied: false,
			},
		},
		{
			Name: "simple_allow",
			Input: []policy.Effect{
				policy.EFFECT_ALLOW,
			},
			Want: output{
				Allowed:          true,
				Denied:           false,
				ExplicitlyDenied: false,
			},
		},
		{
			Name: "simple_deny",
			Input: []policy.Effect{
				policy.EFFECT_DENY,
			},
			Want: output{
				Allowed:          false,
				Denied:           true,
				ExplicitlyDenied: true,
			},
		},

		{
			Name: "explicit_deny",
			Input: []policy.Effect{
				policy.EFFECT_ALLOW,
				policy.EFFECT_DENY,
			},
			Want: output{
				Allowed:          false,
				Denied:           true,
				ExplicitlyDenied: true,
			},
		},
		{
			Name: "many_allows",
			Input: []policy.Effect{
				policy.EFFECT_ALLOW,
				policy.EFFECT_ALLOW,
				policy.EFFECT_ALLOW,
				policy.EFFECT_ALLOW,
				policy.EFFECT_ALLOW,
				policy.EFFECT_ALLOW,
				policy.EFFECT_ALLOW,
			},
			Want: output{
				Allowed:          true,
				Denied:           false,
				ExplicitlyDenied: false,
			},
		},
		{
			Name: "many_denies",
			Input: []policy.Effect{
				policy.EFFECT_DENY,
				policy.EFFECT_DENY,
				policy.EFFECT_DENY,
				policy.EFFECT_DENY,
				policy.EFFECT_DENY,
				policy.EFFECT_DENY,
				policy.EFFECT_DENY,
			},
			Want: output{
				Allowed:          false,
				Denied:           true,
				ExplicitlyDenied: true,
			},
		},
	}

	testlib.RunTestSuite(t, tests, func(e []policy.Effect) (output, error) {
		// Create empty decision
		decision := Decision{}

		// Add our effect rules in
		for _, x := range e {
			decision.Add(x)
		}

		// Ensure size of data never surpasses 2
		if len(decision.Effects()) > 2 {
			t.Fatalf("EffectSet size should never be >2, but saw %d", len(decision.effects))
		}

		return output{
			Allowed:          decision.Allowed(),
			Denied:           decision.Denied(),
			ExplicitlyDenied: decision.ExplicitlyDenied(),
		}, nil
	})
}

func TestMerge(t *testing.T) {
	tests := []testlib.TestCase[[]Decision, []policy.Effect]{
		{
			Name:  "empty",
			Input: []Decision{},
			Want:  []policy.Effect(nil),
		},
		{
			Name: "empties",
			Input: []Decision{
				{},
				{},
				{},
			},
			Want: []policy.Effect(nil),
		},
		{
			Name: "multiple_allows",
			Input: []Decision{
				{effects: []policy.Effect{policy.EFFECT_ALLOW}},
				{effects: []policy.Effect{policy.EFFECT_ALLOW}},
				{effects: []policy.Effect{policy.EFFECT_ALLOW}},
			},
			Want: []policy.Effect{policy.EFFECT_ALLOW},
		},
		{
			Name: "multiple_denies",
			Input: []Decision{
				{effects: []policy.Effect{policy.EFFECT_DENY}},
				{effects: []policy.Effect{policy.EFFECT_DENY}},
				{effects: []policy.Effect{policy.EFFECT_DENY}},
			},
			Want: []policy.Effect{policy.EFFECT_DENY},
		},
		{
			Name: "mix_n_match",
			Input: []Decision{
				{effects: []policy.Effect{policy.EFFECT_ALLOW}},
				{effects: []policy.Effect{policy.EFFECT_DENY}},
				{effects: []policy.Effect{policy.EFFECT_ALLOW, policy.EFFECT_DENY}},
				{effects: []policy.Effect{policy.EFFECT_DENY, policy.EFFECT_ALLOW}},
			},
			Want: []policy.Effect{policy.EFFECT_ALLOW, policy.EFFECT_DENY},
		},
	}

	testlib.RunTestSuite(t, tests, func(d []Decision) ([]policy.Effect, error) {
		// Create empty decision
		decision := Decision{}

		// Perform merges
		for _, x := range d {
			decision.Merge(x)
		}

		// Ensure size of data never surpasses 2
		if len(decision.Effects()) > 2 {
			t.Fatalf("EffectSet size should never be >2, but saw %d", len(decision.effects))
		}

		return decision.Effects(), nil
	})
}
