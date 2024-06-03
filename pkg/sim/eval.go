package sim

import (
	"errors"
	"fmt"
	"slices"

	e "github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/effectset"
	"github.com/nsiow/yams/pkg/sim/gate"
)

// evalOverallAccess calculates both Principal + Resource access same performs both same-account
// and different-account evaluations
func evalOverallAccess(opts *Options, evt *Event) (*Result, error) {

	res := Result{}
	trc := res.Trace

	// Calculate Principal access
	pAccess, err := evalPrincipalAccess(opts, evt, trc)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error evaluating principal access"), err)
	}
	rAccess, err := evalResourceAccess(opts, evt, trc)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error evaluating resource access"), err)
	}

	// Check for explicit Deny results
	if pAccess.Contains(policy.EFFECT_DENY) {
		res.IsAllowed = false
		trc.Add("[explicit deny] explicit deny found in identity policy")
		return &res, nil
	}
	if rAccess.Contains(policy.EFFECT_DENY) {
		res.IsAllowed = false
		trc.Add("[explicit deny] explicit deny found in resource policy")
		return &res, nil
	}

	// If same account, access is granted if the Principal has access
	if evalIsSameAccount(evt.Principal, evt.Resource) {
		if pAccess.Contains(policy.EFFECT_ALLOW) {
			res.IsAllowed = true
			trc.Add("[allow] access granted via same-account identity policy")
			return &res, nil
		}

		// TODO(nsiow) implement correct behavior for same-account access via explicit ARN
		res.IsAllowed = false
		trc.Add("[implicit deny] no identity-based policy allows this action")
		return &res, nil
	}

	// If x-account, access is granted if the Principal has access and the Resource permits that
	// access
	if pAccess.Contains(policy.EFFECT_ALLOW) && rAccess.Contains(policy.EFFECT_ALLOW) {
		res.IsAllowed = true
		trc.Add("[allow] access granted via x-account identity + resource policies")
		return &res, nil
	}
	if pAccess.Contains(policy.EFFECT_ALLOW) && !rAccess.Contains(policy.EFFECT_ALLOW) {
		res.IsAllowed = false
		trc.Add("[implicit deny] x-account, missing resource policy access")
		return &res, nil
	}
	if !pAccess.Contains(policy.EFFECT_ALLOW) && rAccess.Contains(policy.EFFECT_ALLOW) {
		res.IsAllowed = false
		trc.Add("[implicit deny] x-account, missing identity policy access")
		return &res, nil
	}
	res.IsAllowed = false
	trc.Add("[implicit deny] x-account, missing both identity + resource access")
	return &res, nil
}

// statementEvalFunction is the blueprint of a function that allows us to evaluate a single statement
type statementEvalFunction func(*Options, *Event, *Trace, *policy.Statement) (bool, error)

// evalPrincipalAccess calculates the Principal-side access to the specified Resource
func evalPrincipalAccess(opts *Options, evt *Event, trc *Trace) (*effectset.EffectSet, error) {

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
	effects := effectset.EffectSet{}
	for _, polType := range effectivePolicies {
		for _, pol := range polType {
			for _, stmt := range pol.Statement {
				for _, f := range functions {
					match, err := f(opts, evt, trc, &stmt)
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
func evalResourceAccess(opts *Options, evt *Event, trc *Trace) (*effectset.EffectSet, error) {

	// Specify the statement evaluation functions we will consider for Principal access
	functions := []statementEvalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesPrincipal,
		evalStatementMatchesCondition,
	}

	// Iterate over resource policy statements to evaluate access
	effects := effectset.EffectSet{}
	for _, stmt := range evt.Resource.Policy.Statement {
		for _, f := range functions {
			match, err := f(opts, evt, trc, &stmt)
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
func evalStatementMatchesAction(
	opts *Options, evt *Event, trc *Trace, stmt *policy.Statement) (bool, error) {

	// Determine which Action block to use
	var _gate gate.Gate
	var action policy.Action
	if !stmt.Action.Empty() {
		action = stmt.Action
	} else {
		action = stmt.NotAction
		_gate.Invert()
	}

	for _, a := range action {
		match := matchWildcardIgnoreCase(a, evt.Action)
		if match {
			return _gate.Apply(true), nil
		}
	}

	return _gate.Apply(false), nil
}

// evalStatementMatchesPrincipal computes whether the Statement matches the Event's Principal
func evalStatementMatchesPrincipal(
	opts *Options, evt *Event, trc *Trace, stmt *policy.Statement) (bool, error) {

	// Determine which Principal block to use
	var _gate gate.Gate
	var principals policy.Principal
	if !stmt.Principal.Empty() {
		principals = stmt.Principal
	} else {
		principals = stmt.NotPrincipal
		_gate.Invert()
	}

	// TODO(nsiow) this may need to change for subresource based operations e.g. s3:getobject
	for _, p := range principals.AWS {
		match := matchWildcard(p, evt.Principal.Arn)
		if match {
			return _gate.Apply(true), nil
		}
	}

	return _gate.Apply(false), nil
}

// evalStatementMatchesResource computes whether the Statement matches the Event's Resource
func evalStatementMatchesResource(
	opts *Options, evt *Event, trc *Trace, stmt *policy.Statement) (bool, error) {

	// Determine which Resource block to use
	var _gate gate.Gate
	var resources policy.Resource
	if !stmt.Resource.Empty() {
		resources = stmt.Resource
	} else {
		resources = stmt.NotResource
		_gate.Invert()
	}

	// FIXME(nsiow) is this how gate should be used?
	for _, r := range resources {
		match := matchWildcard(r, evt.Resource.Arn)
		if match {
			return _gate.Apply(true), nil
		}
	}

	return _gate.Apply(false), nil
}

// evalStatementMatchesCondition computes whether the Statement's Conditions hold true given the
// provided Event
func evalStatementMatchesCondition(
	opts *Options, evt *Event, trc *Trace, stmt *policy.Statement) (bool, error) {
	// FIXME(nsiow) this needs to be implemented

	knownConditionOperators := []string{}
	for op := range stmt.Condition {
		if !slices.Contains(knownConditionOperators, op) {
			return false, fmt.Errorf("unknown condition operator: %s", op)
		}
	}

	return true, nil
}

// evalIsSameAccount determines whether or not the provided Principal + Resource exist within the
// same AWS account
func evalIsSameAccount(p *e.Principal, r *e.Resource) bool {
	return p.Account == r.Account
}
