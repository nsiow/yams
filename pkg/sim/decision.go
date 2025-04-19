package sim

import (
	"slices"

	"github.com/nsiow/yams/pkg/policy"
)

// TODO(nsiow) consider renaming to Decision

// Decision maintains a unique list of Effect values
type Decision struct {
	effects []policy.Effect
}

// Add takes the provided Effect and saves it to the Decision if it is a new value
func (e *Decision) Add(effect policy.Effect) {
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
func (e *Decision) Effects() []policy.Effect {
	return e.effects
}

// Contains determines whether or not the specified Effect is present in our set
func (e *Decision) Contains(effect policy.Effect) bool {
	return slices.Contains(e.effects, effect)
}

// Allowed determines whether or not the Decision corresponds to an IAM operation being allowed,
// based on the values contained within the set
func (e *Decision) Allowed() bool {
	return e.Contains(policy.EFFECT_ALLOW) && !e.Contains(policy.EFFECT_DENY)
}

// Denied determines whether or not the Decision corresponds to an IAM operation being denied
// based on the values contained within the set
func (e *Decision) Denied() bool {
	return e.Contains(policy.EFFECT_DENY) || !e.Contains(policy.EFFECT_ALLOW)
}

// ExplicitlyDenied determines whether or not the Decision corresponds to an IAM operation being
// denied based on an explicit DENY decision contained within the set
// TODO(nsiow) check for other instances in the code base where this should be used
func (e *Decision) ExplicitlyDenied() bool {
	return e.Contains(policy.EFFECT_DENY)
}

// Merge combines the provided Decision(s) with our target
func (e *Decision) Merge(others ...Decision) {
	for _, other := range others {
		for _, effect := range other.Effects() {
			e.Add(effect)
		}
	}
}
