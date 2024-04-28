package entities

import (
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/tag"
)

// Principal defines the general shape of an AWS cloud principal
type Principal struct {
	Type    string
	Account string
	Region  string
	Arn     string
	Tags    []tag.Tag
	// FIXME(nsiow) this isn't really the right shape
	InlinePolicies  []policy.Statement
	ManagedPolicies []policy.Statement
}
