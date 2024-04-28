package entities

import (
	"github.com/nsiow/yams/pkg/policy"
)

// Principal defines the general shape of an AWS cloud principal
type Principal struct {
	Type    string
	Account string
	Region  string
	Arn     string
	Tags    []Tag
	// FIXME(nsiow) this isn't really the right shape
	InlinePolicies  []policy.Policy
	ManagedPolicies []policy.Policy
}
