package entities

import (
	"iter"
	"maps"

	"github.com/nsiow/yams/pkg/policy"
)

// Universe contains the definition of all accounts/policies/principals/resources used for
// simulation.
//
// In order for something to be considered for simulation, it must be a part of this universe
type Universe struct {
	accounts   map[Arn]Account
	policies   map[Arn]policy.Policy
	principals map[Arn]Principal
	resources  map[Arn]Resource
}

// Accounts returns an iterator over all the Account entities known to the universe
func (u *Universe) Accounts() iter.Seq[Account] {
	return maps.Values(u.accounts)
}

// Policies returns an iterator over all the IAM policies known to the universe
//
// This includes any policy with an ARN, e.g. managed policies, SCPs, etc. It does not include
// inline Principal or Resource policies
func (u *Universe) Policies() iter.Seq[policy.Policy] {
	return maps.Values(u.policies)
}

// Principals returns an iterator over all the Principal entities known to the universe
func (u *Universe) Principals() iter.Seq[Principal] {
	return maps.Values(u.principals)
}

// Resources returns an iterator over all the Resource entities known to the universe
func (u *Universe) Resources() iter.Seq[Resource] {
	return maps.Values(u.resources)
}
