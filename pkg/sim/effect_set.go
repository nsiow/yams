package sim

import (
	"slices"

	"github.com/nsiow/yams/pkg/policy"
)

// EffectSet maintains a unique list of Effect values
type EffectSet struct {
	effects []policy.Effect
}

// Add takes the provided Effect and saves it to the EffectSet if it is a new value
func (e *EffectSet) Add(effect policy.Effect) {
	if !slices.Contains(e.effects, effect) {
		e.effects = append(e.effects, effect)
	}
}

// Contains determines whether or not the specified Effect is present in our set
func (e *EffectSet) Contains(effect policy.Effect) bool {
	return slices.Contains(e.effects, effect)
}

// Allowed determines whether or not the EffectSet corresponds to an IAM operation being allowed,
// based on the values contained within the set
func (e *EffectSet) Allowed() bool {
	return len(e.effects) == 1 && e.effects[0] == policy.EFFECT_ALLOW
}
