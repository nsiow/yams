package sim

import (
	"testing"

	"github.com/nsiow/yams/pkg/policy"
)

func TestEffectSet(t *testing.T) {
	type test struct {
		name    string
		input   []policy.Effect
		allowed bool
	}

	tests := []test{
		{
			name:    "implicit_deny",
			input:   []policy.Effect{},
			allowed: false,
		},
		{
			name: "simple_allow",
			input: []policy.Effect{
				policy.EFFECT_ALLOW,
			},
			allowed: true,
		},
		{
			name: "simple_deny",
			input: []policy.Effect{
				policy.EFFECT_DENY,
			},
			allowed: false,
		},

		{
			name: "explicit_deny",
			input: []policy.Effect{
				policy.EFFECT_ALLOW,
				policy.EFFECT_DENY,
			},
			allowed: false,
		},
		{
			name: "many_allows",
			input: []policy.Effect{
				policy.EFFECT_ALLOW,
				policy.EFFECT_ALLOW,
				policy.EFFECT_ALLOW,
				policy.EFFECT_ALLOW,
				policy.EFFECT_ALLOW,
				policy.EFFECT_ALLOW,
				policy.EFFECT_ALLOW,
			},
			allowed: true,
		},
		{
			name: "many_denies",
			input: []policy.Effect{
				policy.EFFECT_DENY,
				policy.EFFECT_DENY,
				policy.EFFECT_DENY,
				policy.EFFECT_DENY,
				policy.EFFECT_DENY,
				policy.EFFECT_DENY,
				policy.EFFECT_DENY,
			},
			allowed: false,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		// Add effects to effect set
		es := EffectSet{}
		for _, effect := range tc.input {
			es.Add(effect)
		}

		// Ensure that we never go over 2 elements
		if len(es.effects) > 2 {
			t.Fatalf("saw %d elements in EffectSet; expected a maximum of 2", len(es.effects))
		}

		// Check results
		got := es.Allowed()
		if tc.allowed != got {
			t.Fatalf("failed test case '%s': wanted %v got %v", tc.name, tc.allowed, es.Allowed())
		}
	}
}
