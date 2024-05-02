package entities

import (
	"github.com/nsiow/yams/pkg/policy"
)

// Principal defines the general shape of an AWS cloud principal
type Principal struct {
	Type            string
	Account         string
	Region          string
	Arn             string
	Tags            []Tag
	InlinePolicies  []policy.Policy
	ManagedPolicies []policy.Policy
}

// Policies returns the merger of both inline and managed policies
func (p *Principal) Policies() []policy.Policy {
	// Do it the old fashioned way rather than `append` to avoid recopies
	length := len(p.InlinePolicies) + len(p.ManagedPolicies)
	policies := make([]policy.Policy, length)

	i := 0
	for j := range p.InlinePolicies {
		policies[i] = p.InlinePolicies[j]
		i += 1
	}
	for k := range p.ManagedPolicies {
		policies[i] = p.ManagedPolicies[k]
		i += 1
	}

	return policies
}

// Statements returns the merger of statements within both inline and managed policies
func (p *Principal) Statements() []policy.Statement {
	var stmts []policy.Statement

	// Iterate over policies => statement blocks => statements
	for _, policy := range p.Policies() {
		for _, stmt := range policy.Statement {
			stmts = append(stmts, stmt)
		}
	}
	return stmts
}
