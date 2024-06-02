package sim

import (
	"errors"
	"fmt"

	e "github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// evalIsSameAccount determines whether or not the provided Principal + Resource exist within the
// same AWS account
func evalIsSameAccount(p *e.Principal, r *e.Resource) bool {
	return p.Account == r.Account
}

// evalCombinedAccess calculates the Principal + Resource access

// evalPrincipalAccess calculates the Principal-side access to the specified Resource
func evalPrincipalAccess(action string,
	p *e.Principal,
	r *e.Resource,
	ac *AuthContext) (*EffectSet, error) {

	// Specify the types of policies we will consider for Principal access
	effectivePolicies := [][]policy.Policy{
		p.InlinePolicies,
		p.AttachedPolicies,
	}

	// Iterate over policy types / policies / statements to evaluate access
	effects := EffectSet{}
	for _, polType := range effectivePolicies {
		for _, pol := range polType {
			for _, stmt := range pol.Statement {
				matches, err := evalStatementMatchesResource(stmt, p, r, ac)
				if err != nil {
					return nil, errors.Join(
						fmt.Errorf("error evaluating principal policy statement[sid=%s]", stmt.Sid),
						err)
				}
				if matches {
					effects.Add(stmt.Effect)
				}
			}
		}
	}

	return &effects, nil
}

// evalResourceAccess calculates the Resource-side access with regard to the specified Principal
func evalResourceAccess(action string,
	p *e.Principal,
	r *e.Resource,
	ac *AuthContext) (*EffectSet, error) {

	// Iterate over policy types / policies / statements to evaluate access
	effects := EffectSet{}
	for _, stmt := range r.Policy.Statement {
		matches, err := evalStatementMatchesPrincipal(stmt, p, r, ac)
		if err != nil {
			return nil, errors.Join(
				fmt.Errorf("error evaluating resource policy statement[sid=%s]", stmt.Sid),
				err)
		}
		if matches {
			effects.Add(stmt.Effect)
		}
	}

	return &effects, nil

}

func evalStatementMatchesPrincipal(stmt policy.Statement,
	p *e.Principal,
	r *e.Resource,
	ac *AuthContext) (bool, error) {
	return true, nil
}

func evalStatementMatchesResource(stmt policy.Statement,
	p *e.Principal,
	r *e.Resource,
	ac *AuthContext) (bool, error) {
	return true, nil
}
