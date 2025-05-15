package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
)

// Decision maintains a unique list of Effect values
type Decision struct {
	allow bool
	deny  bool
}

// Add takes the provided Effect and saves it to the Decision
func (d *Decision) Add(effect policy.Effect) {
	switch effect {
	case policy.EFFECT_ALLOW:
		d.allow = true
	case policy.EFFECT_DENY:
		d.deny = true
	default:
		panic(fmt.Sprintf("wtf is %s", effect))
	}
}

// Allowed determines whether or not the Decision corresponds to an IAM operation being allowed
func (d *Decision) Allowed() bool {
	return d.allow && !d.deny
}

// Denied determines whether or not the Decision corresponds to an IAM operation being denied
func (d *Decision) Denied() bool {
	return d.deny || !d.allow
}

// DeniedExplicit determines whether or not the Decision corresponds to an IAM operation being
// denied based on an explicit DENY decision
// TODO(nsiow) check for other instances in the code base where this should be used
func (d *Decision) DeniedExplicit() bool {
	return d.deny
}

// Merge combines the provided Decision(s) with our target
func (d *Decision) Merge(others ...Decision) {
	for _, other := range others {
		d.allow = d.allow || other.allow
		d.deny = d.deny || other.deny
	}
}
