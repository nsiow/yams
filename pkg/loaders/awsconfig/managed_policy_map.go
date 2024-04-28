package awsconfig

import "github.com/nsiow/yams/pkg/policy"

// ManagedPolicyMap contains a mapping from ARN to policy for AWS- and customer-managed policies
type ManagedPolicyMap struct {
	pmap map[string]policy.Policy
}

// NewManagedPolicyMap creates and returns an initialized instance of ManagedPolicyMap
func NewManagedPolicyMap() *ManagedPolicyMap {
	m := ManagedPolicyMap{}
	m.pmap := make(map[string]policy.Policy)
	return &m
}

// Add creates a new mapping between the provided ARN and policy
func (m *ManagedPolicyMap) Add(arn string, pstruct policy.Policy) {
	m.pmap[arn] = pstruct
}

// Get retrieves the requested policy by ARN, if it exists
func (m *ManagedPolicyMap) Get(arn string) (policy.Policy, bool) {
	val, ok := m.pmap[arn]
	return val, ok
}
