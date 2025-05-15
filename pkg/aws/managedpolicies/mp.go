package managedpolicies

import (
	"github.com/nsiow/yams/internal/assets"
	"github.com/nsiow/yams/pkg/policy"
)

// data is a local alias hiding the asset implementation of MP data
var data = assets.ManagedPolicyData

// Map returns a map with format key=arn, value=policy for all managed policies
func Map() map[string]policy.Policy {
	return data()
}

// All returns a slice containing all the known managed policies
func All() []policy.Policy {
	data := data()
	policies := make([]policy.Policy, len(data))

	for _, policy := range data {
		policies = append(policies, policy)
	}

	return policies
}

// Get returns the requested managed policy based on the provided ARN
//
// The second value is true if the policy was successfully found and false otherwise
func Get(arn string) (policy.Policy, bool) {
	p, ok := data()[arn]
	return p, ok
}
