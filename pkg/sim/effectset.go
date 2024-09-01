package sim

import (
	"slices"

	"github.com/nsiow/yams/pkg/policy"
)

// TODO(nsiow) consider renaming to Decision

// EffectSet maintains a unique list of Effect values
type EffectSet struct {
	effects []policy.Effect
}

// Add takes the provided Effect and saves it to the EffectSet if it is a new value
func (e *EffectSet) Add(effect policy.Effect) {
	if !slices.Contains(e.effects, effect) {
		// Change insertion point based on effect, so that ordering is always consistent
		if effect == policy.EFFECT_ALLOW {
			e.effects = slices.Insert(e.effects, 0, effect)
		} else {
			e.effects = append(e.effects, effect)
		}
	}
}

// Effects returns all Effect values currently held within the set
func (e *EffectSet) Effects() []policy.Effect {
	return e.effects
}

// Contains determines whether or not the specified Effect is present in our set
func (e *EffectSet) Contains(effect policy.Effect) bool {
	return slices.Contains(e.effects, effect)
}

// Allowed determines whether or not the EffectSet corresponds to an IAM operation being allowed,
// based on the values contained within the set
func (e *EffectSet) Allowed() bool {
	return e.Contains(policy.EFFECT_ALLOW) && !e.Contains(policy.EFFECT_DENY)
}

// Denied determines whether or not the EffectSet corresponds to an IAM operation being denied
// based on the values contained within the set
func (e *EffectSet) Denied() bool {
	return e.Contains(policy.EFFECT_DENY) || !e.Contains(policy.EFFECT_ALLOW)
}

// ExplicitlyDenied determines whether or not the EffectSet corresponds to an IAM operation being
// denied based on an explicit DENY decision contained within the set
// TODO(nsiow) check for other instances in the code base where this should be used
func (e *EffectSet) ExplicitlyDenied() bool {
	return e.Contains(policy.EFFECT_DENY)
}

// Merge combines the provided EffectSet with our target
func (e *EffectSet) Merge(other EffectSet) {
	for _, effect := range other.Effects() {
		e.Add(effect)
	}
}
