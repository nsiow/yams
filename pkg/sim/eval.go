package sim

import (
	"errors"
	"fmt"

	e "github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// evalOverallAccess calculates both Principal + Resource access same performs both same-account
// and different-account evaluations
func evalOverallAccess(evt *Event) (*Result, error) {

	res := Result{}

	// Calculate Principal access
	pAccess, err := evalPrincipalAccess(evt)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error evaluating principal access"), err)
	}
	rAccess, err := evalResourceAccess(evt)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error evaluating resource access"), err)
	}

	// Check for explicit Deny results
	if pAccess.Contains(policy.EFFECT_DENY) {
		res.IsAllowed = false
		res.ResultContext.Add("[explicit deny] explicit deny found in identity policy")
		return &res, nil
	}
	if rAccess.Contains(policy.EFFECT_DENY) {
		res.IsAllowed = false
		res.ResultContext.Add("[explicit deny] explicit deny found in resource policy")
		return &res, nil
	}

	// If same account, access is granted if the Principal has access
	if evalIsSameAccount(evt.Principal, evt.Resource) {
		if pAccess.Contains(policy.EFFECT_ALLOW) {
			res.IsAllowed = true
			res.ResultContext.Add("[allow] access granted via same-account identity policy")
			return &res, nil
		}

		// TODO(nsiow) implement correct behavior for same-account access via explicit ARN
		res.IsAllowed = false
		res.ResultContext.Add("[implicit deny] no identity-based policy allows this action")
		return &res, nil
	}

	// If x-account, access is granted if the Principal has access and the Resource permits that
	// access
	if pAccess.Contains(policy.EFFECT_ALLOW) && rAccess.Contains(policy.EFFECT_ALLOW) {
		res.IsAllowed = true
		res.ResultContext.Add("[allow] access granted via x-account identity + resource policies")
		return &res, nil
	}
	if pAccess.Contains(policy.EFFECT_ALLOW) && !rAccess.Contains(policy.EFFECT_ALLOW) {
		res.IsAllowed = false
		res.ResultContext.Add("[implicit deny] x-account, missing resource policy access")
		return &res, nil
	}
	if !pAccess.Contains(policy.EFFECT_ALLOW) && rAccess.Contains(policy.EFFECT_ALLOW) {
		res.IsAllowed = false
		res.ResultContext.Add("[implicit deny] x-account, missing identity policy access")
		return &res, nil
	}
	res.IsAllowed = false
	res.ResultContext.Add("[implicit deny] x-account, missing both identity + resource access")
	return &res, nil
}

// statementEvalFunction is the blueprint of a function that allows us to evaluate a single statement
type statementEvalFunction func(*Event, *policy.Statement) (bool, error)

// evalPrincipalAccess calculates the Principal-side access to the specified Resource
func evalPrincipalAccess(evt *Event) (*EffectSet, error) {

	// Specify the types of policies we will consider for Principal access
	effectivePolicies := [][]policy.Policy{
		evt.Principal.InlinePolicies,
		evt.Principal.AttachedPolicies,
	}

	// Specify the statement evaluation functions we will consider for Principal access
	functions := []statementEvalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesCondition,
	}

	// Iterate over policy types / policies / statements to evaluate access
	effects := EffectSet{}
	for _, polType := range effectivePolicies {
		for _, pol := range polType {
			for _, stmt := range pol.Statement {
				for _, f := range functions {
					match, err := f(evt, &stmt)
					if err != nil {
						return nil, errors.Join(
							fmt.Errorf("error evaluating principal policy statement[sid=%s]", stmt.Sid),
							err)
					}
					if match {
						effects.Add(stmt.Effect)
					}
				}
			}
		}
	}

	return &effects, nil
}

// evalResourceAccess calculates the Resource-side access with regard to the specified Principal
func evalResourceAccess(evt *Event) (*EffectSet, error) {

	// Specify the statement evaluation functions we will consider for Principal access
	functions := []statementEvalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesPrincipal,
		evalStatementMatchesCondition,
	}

	// Iterate over policy types / policies / statements to evaluate access
	effects := EffectSet{}
	for _, stmt := range evt.Resource.Policy.Statement {
		for _, f := range functions {
			match, err := f(evt, &stmt)
			if err != nil {
				return nil, errors.Join(
					fmt.Errorf("error evaluating principal policy statement[sid=%s]", stmt.Sid),
					err)
			}
			if match {
				effects.Add(stmt.Effect)
			}
		}
	}

	return &effects, nil

}

// evalStatementMatchesAction computes whether the Statement matches the Event's Action
func evalStatementMatchesAction(evt *Event, stmt *policy.Statement) (bool, error) {
	panic("not yet implemented")
}

// evalStatementMatchesCondition computes whether the Statement's Conditions hold true given the
// provided Event
func evalStatementMatchesCondition(evt *Event, stmt *policy.Statement) (bool, error) {
	panic("not yet implemented")
}

// evalStatementMatchesPrincipal computes whether the Statement matches the Event's Principal
func evalStatementMatchesPrincipal(evt *Event, stmt *policy.Statement) (bool, error) {
	panic("not yet implemented")
}

// evalStatementMatchesResource computes whether the Statement matches the Event's Resource
func evalStatementMatchesResource(evt *Event, stmt *policy.Statement) (bool, error) {
	panic("not yet implemented")
}

// evalIsSameAccount determines whether or not the provided Principal + Resource exist within the
// same AWS account
func evalIsSameAccount(p *e.Principal, r *e.Resource) bool {
	return p.Account == r.Account
}
