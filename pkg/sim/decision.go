package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
)

// Decision maintains a unique list of Effect values
type Decision struct {
	Allow bool
	Deny  bool
}

// Add takes the provided Effect and saves it to the Decision
func (d *Decision) Add(effect policy.Effect) {
	switch effect {
	case policy.EFFECT_ALLOW:
		d.Allow = true
	case policy.EFFECT_DENY:
		d.Deny = true
	default:
		panic(fmt.Sprintf("wtf is %s", effect))
	}
}

// Allowed determines whether or not the Decision corresponds to an IAM operation being allowed
func (d *Decision) Allowed() bool {
	return d.Allow && !d.Deny
}

// Denied determines whether or not the Decision corresponds to an IAM operation being denied
func (d *Decision) Denied() bool {
	return d.Deny || !d.Allow
}

// DeniedExplicit determines whether or not the Decision corresponds to an IAM operation being
// denied based on an explicit DENY decision
// TODO(nsiow) check for other instances in the code base where this should be used
func (d *Decision) DeniedExplicit() bool {
	return d.Deny
}

// Merge combines the provided Decision(s) with our target
func (d *Decision) Merge(others ...Decision) {
	for _, other := range others {
		d.Allow = d.Allow || other.Allow
		d.Deny = d.Deny || other.Deny
	}
}
