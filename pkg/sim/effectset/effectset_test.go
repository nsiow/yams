package effectset

import (
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
	"github.com/nsiow/yams/pkg/policy"
)

func TestEffectSet(t *testing.T) {
	tests := []testrunner.TestCase[[]policy.Effect, bool]{
		{
			Name:  "implicit_deny",
			Input: []policy.Effect{},
			Want:  false,
		},
		{
			Name: "simple_allow",
			Input: []policy.Effect{
				policy.EFFECT_ALLOW,
			},
			Want: true,
		},
		{
			Name: "simple_deny",
			Input: []policy.Effect{
				policy.EFFECT_DENY,
			},
			Want: false,
		},

		{
			Name: "explicit_deny",
			Input: []policy.Effect{
				policy.EFFECT_ALLOW,
				policy.EFFECT_DENY,
			},
			Want: false,
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
			Want: true,
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
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(e []policy.Effect) (bool, error) {
		// Create empty effectset
		es := EffectSet{}

		// Add our effect rules in
		for _, i := range e {
			es.Add(i)
		}

		// Ensure size of data never surpasses 2
		if len(es.Effects()) > 2 {
			t.Fatalf("EffectSet size should never be >2, but saw %d", len(es.effects))
		}

		return es.Allowed(), nil
	})
}
