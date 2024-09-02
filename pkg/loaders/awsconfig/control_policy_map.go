package awsconfig

import (
	"github.com/nsiow/yams/pkg/policy"
)

// ControlPolicyMap contains a mapping from Account to SCPs
type ControlPolicyMap struct {
	mapping map[string][][]policy.Policy
}

// NewControlPolicyMap creates and returns an initialized instance of ControlPolicyMap
func NewControlPolicyMap() *ControlPolicyMap {
	m := ControlPolicyMap{}
	m.mapping = make(map[string][][]policy.Policy)
	return &m
}

// Add creates a new mapping between the provided Account and SCPs
func (m *ControlPolicyMap) Add(account string, pstruct [][]policy.Policy) {
	m.mapping[account] = pstruct
}

// Get retrieves the requested SCPs by Account, if it exists
// TODO(nsiow) figure out if SCPs should fail open or fail closed if the account is not found...
// simulator configuration? based on whether or not any SCPs have been given? what's the least
// surprising behavior
func (m *ControlPolicyMap) Get(account string) ([][]policy.Policy, bool) {
	val, ok := m.mapping[account]
	return val, ok
}
