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

		return output{
			Allowed:          decision.Allowed(),
			Denied:           decision.Denied(),
			ExplicitlyDenied: decision.DeniedExplicit(),
		}, nil
	})
}

func TestMerge(t *testing.T) {
	tests := []testlib.TestCase[[]Decision, Decision]{
		{
			Name:  "empty",
			Input: []Decision{},
			Want:  Decision{},
		},
		{
			Name: "empties",
			Input: []Decision{
				{},
				{},
				{},
			},
			Want: Decision{},
		},
		{
			Name: "multiple_allows",
			Input: []Decision{
				Decision{Allow: true},
				Decision{Allow: true},
				Decision{Allow: true},
			},
			Want: Decision{Allow: true},
		},
		{
			Name: "multiple_denies",
			Input: []Decision{
				Decision{Deny: true},
				Decision{Deny: true},
				Decision{Deny: true},
			},
			Want: Decision{Deny: true},
		},
		{
			Name: "mix_n_match",
			Input: []Decision{
				Decision{Allow: true},
				Decision{Deny: true},
				Decision{Allow: true, Deny: true},
				Decision{Deny: true},
			},
			Want: Decision{Allow: true, Deny: true},
		},
	}

	testlib.RunTestSuite(t, tests, func(d []Decision) (Decision, error) {
		// Create empty decision
		decision := Decision{}

		// Perform merges
		for _, x := range d {
			decision.Merge(x)
		}

		return decision, nil
	})
}
