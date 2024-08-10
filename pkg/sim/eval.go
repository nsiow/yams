package sim

import (
	"errors"
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/gate"
	"github.com/nsiow/yams/pkg/sim/trace"
	"github.com/nsiow/yams/pkg/sim/wildcard"
)

// evalIsSameAccount determines whether or not the provided Principal + Resource exist within the
// same AWS account
func evalIsSameAccount(p *entities.Principal, r *entities.Resource) bool {
	return p.Account == r.Account
}

// evalOverallAccess calculates both Principal + Resource access same performs both same-account
// and different-account evaluations
func evalOverallAccess(opt *Options, ac AuthContext) (*Result, error) {

	trc := trace.New()

	trc.Attr("authContext", ac)

	// Calculate permissions boundary access, if present
	pbAccess, err := evalPermissionsBoundary(trc, opt, ac)
	if err != nil {
		return nil, fmt.Errorf("error evaluating permission boundary: %w", err)
	}
	if pbAccess.Contains(policy.EFFECT_DENY) {
		trc.Decision("[explicit deny] found in permissions boundary")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}

	// Calculate Principal access
	pAccess, err := evalPrincipalAccess(trc, opt, ac)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error evaluating principal access"), err)
	}
	// ... check for explicit Deny results
	if pAccess.Contains(policy.EFFECT_DENY) {
		trc.Decision("[explicit deny] found in identity policy")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}

	// Calculate Resource access
	rAccess, err := evalResourceAccess(trc, opt, ac)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error evaluating resource access"), err)
	}
	// ... check for explicit Deny results
	if rAccess.Contains(policy.EFFECT_DENY) {
		trc.Decision("[explicit deny] found in resource policy")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}

	// If same account, access is granted if the Principal has access
	if evalIsSameAccount(ac.Principal, ac.Resource) {
		if pAccess.Contains(policy.EFFECT_ALLOW) {
			trc.Decision("[allow] access granted via same-account identity policy")
			return &Result{Trace: trc, IsAllowed: true}, nil
		}

		// TODO(nsiow) implement correct behavior for same-account access via explicit ARN
		trc.Decision("[implicit deny] no identity-based policy allows this action")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}

	// If x-account, access is granted if the Principal has access and the Resource permits that
	// access
	if pAccess.Contains(policy.EFFECT_ALLOW) && rAccess.Contains(policy.EFFECT_ALLOW) {
		trc.Decision("[allow] access granted via x-account identity + resource policies")
		return &Result{Trace: trc, IsAllowed: true}, nil
	}
	if pAccess.Contains(policy.EFFECT_ALLOW) && !rAccess.Contains(policy.EFFECT_ALLOW) {
		trc.Decision("[implicit deny] x-account, missing resource policy access")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}
	if !pAccess.Contains(policy.EFFECT_ALLOW) && rAccess.Contains(policy.EFFECT_ALLOW) {
		trc.Decision("[implicit deny] x-account, missing identity policy access")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}

	// We fell through and no access was granted from either side
	trc.Decision("[implicit deny] x-account, missing both identity + resource access")
	return &Result{Trace: trc, IsAllowed: false}, nil
}

// evalFunction is the blueprint of a function that allows us to evaluate a single statement
type evalFunction func(*trace.Trace, *Options, AuthContext, *policy.Statement) (bool, error)

// evalPrincipalAccess calculates the Principal-side access to the specified Resource
func evalPrincipalAccess(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext) (EffectSet, error) {

	trc.Push("evaluating principal policies")
	defer trc.Pop()

	// Specify the types of policies we will consider for Principal access
	effectivePolicies := map[string][]policy.Policy{
		"inline":  ac.Principal.InlinePolicies,
		"managed": ac.Principal.AttachedPolicies,
		"group":   ac.Principal.GroupPolicies,
	}

	// Specify the statement evaluation funcs we will consider for Principal access
	funcs := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesCondition,
	}

	// Iterate over policy types / policies / statements to evaluate access
	effects := EffectSet{}
	for policytype, policies := range effectivePolicies {
		trc.Push(fmt.Sprintf("policytype=%s", policytype))
		eff, err := evalPolicies(trc, opt, ac, policies, funcs)
		if err != nil {
			trc.Pop()
			return effects, err
		}

		effects.Merge(eff)
		trc.Pop()
	}

	return effects, nil
}

// evalResourceAccess calculates the Resource-side access with regard to the specified Principal
func evalResourceAccess(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext) (EffectSet, error) {

	trc.Push("evaluating resource policies")
	defer trc.Pop()

	// Specify the statement evaluation funcs we will consider for Principal access
	funcs := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesPrincipal,
		evalStatementMatchesCondition,
	}

	// Iterate over resource policy statements to evaluate access
	return evalPolicy(trc, opt, ac, ac.Resource.Policy, funcs)
}

// evalPolicies computes whether the provided policies match the AuthContext
func evalPolicies(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	policies []policy.Policy,
	funcs []evalFunction) (EffectSet, error) {

	effects := EffectSet{}

	for _, pol := range policies {
		eff, err := evalPolicy(trc, opt, ac, pol, funcs)
		if err != nil {
			return eff, err
		}

		// TODO(nsiow) short circuit on Deny, or keep going for completeness?

		effects.Merge(eff)
	}

	return effects, nil
}

// evalPolicy computes whether the provided policy matches the AuthContext
// FIXME(nsiow) re-add trace statements to all of the below functions
// (evalPolicy/evalPolicies/evalStatement)
func evalPolicy(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	policy policy.Policy,
	funcs []evalFunction) (EffectSet, error) {

	effects := EffectSet{}

	for _, stmt := range policy.Statement {
		eff, err := evalStatement(trc, opt, ac, stmt, funcs)
		if err != nil {
			return eff, err
		}

		effects.Merge(eff)
	}

	return effects, nil
}

// evalStatement computes whether the provided statements match the AuthContext
func evalStatement(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	stmt policy.Statement,
	funcs []evalFunction) (EffectSet, error) {

	for _, f := range funcs {
		match, err := f(trc, opt, ac, &stmt)
		if err != nil {
			return EffectSet{}, err
		}

		if !match {
			return EffectSet{}, nil
		}
	}

	effects := EffectSet{}
	effects.Add(stmt.Effect)
	return effects, nil
}

// evalStatementMatchesAction computes whether the Statement matches the AuthContext's Action
func evalStatementMatchesAction(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	stmt *policy.Statement) (bool, error) {

	trc.Push("evaluating Action")
	defer trc.Pop()

	// Handle empty Action
	if len(ac.Action) == 0 {
		trc.Observation("AuthContext missing Action")
		return false, nil
	}

	// Determine which Action block to use
	var _gate gate.Gate
	var action policy.Action
	if !stmt.Action.Empty() {
		trc.Observation("using Action block")
		action = stmt.Action
	} else {
		trc.Observation("using NotAction block")
		action = stmt.NotAction
		_gate.Invert()
	}

	for _, a := range action {
		match := wildcard.MatchSegmentsIgnoreCase(a, ac.Action)
		if match {
			trc.Attr("action", a)
			trc.Observation("action matched")
			return _gate.Apply(true), nil
		}
	}

	trc.Observation("no actions match")
	return _gate.Apply(false), nil
}

// evalStatementMatchesPrincipal computes whether the Statement matches the AuthContext's Principal
func evalStatementMatchesPrincipal(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	stmt *policy.Statement) (bool, error) {

	trc.Push("evaluating Principal")
	defer trc.Pop()

	// Handle empty Principal
	if ac.Principal == nil {
		trc.Observation("AuthContext missing Principal")
		return false, nil
	}

	// Determine which Principal block to use
	var _gate gate.Gate
	var principals policy.Principal
	if !stmt.Principal.Empty() {
		trc.Observation("using Principal block")
		principals = stmt.Principal
	} else {
		trc.Observation("using NotPrincipal block")
		principals = stmt.NotPrincipal
		_gate.Invert()
	}

	// TODO(nsiow) validate that this is how Principals are evaluated - exact matches?
	for _, p := range principals.AWS {
		match := wildcard.MatchAllOrNothing(p, ac.Principal.Arn)
		if match {
			return _gate.Apply(true), nil
		}
	}

	return _gate.Apply(false), nil
}

// evalStatementMatchesResource computes whether the Statement matches the AuthContext's Resource
func evalStatementMatchesResource(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	stmt *policy.Statement) (bool, error) {

	trc.Push("evaluating Resource")
	defer trc.Pop()

	// Handle empty Resource
	if ac.Resource == nil {
		trc.Observation("AuthContext missing Resource")
		return false, nil
	}

	// Determine which Resource block to use
	var _gate gate.Gate
	var resources policy.Resource
	if !stmt.Resource.Empty() {
		trc.Observation("using Resource block")
		resources = stmt.Resource
	} else {
		trc.Observation("using NotResource block")
		resources = stmt.NotResource
		_gate.Invert()
	}

	// TODO(nsiow) this may need to change for subresource based operations e.g. s3:getobject
	// TODO(nsiow) this needs to support variable expansion
	for _, r := range resources {
		match := wildcard.MatchSegments(r, ac.Resource.Arn)
		if match {
			return _gate.Apply(true), nil
		}
	}

	return _gate.Apply(false), nil
}

// evalStatementMatchesCondition computes whether the Statement's Conditions hold true given the
// provided AuthContext
func evalStatementMatchesCondition(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	stmt *policy.Statement) (bool, error) {

	trc.Push("evaluating Condition")
	defer trc.Pop()

	for op, cond := range stmt.Condition {
		result, err := evalCheckCondition(trc, opt, ac, op, cond)
		if err != nil {
			return false, err
		}

		if !result {
			trc.Observation("condition evaluated to false")
			return false, nil
		}
	}

	trc.Observation("condition evaluated to true")
	return true, nil
}

// evalCheckCondition assesses a single condition operator to determine whether or not it matches
// the provided AuthContext
func evalCheckCondition(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	op string,
	cond map[string]policy.Value) (bool, error) {

	// TODO(nsiow) implement PushWithAttr so that `op` is in a more appropriate context?
	trc.Push("evaluating Operation")
	trc.Attr("op", op)
	defer trc.Pop()

	// An empty condition should actually evaluate to false
	if len(cond) == 0 {
		trc.Observation("no match; empty condition")
		return false, nil
	}

	// Check to see if the condition operator is supported
	f, exists := ResolveConditionEvaluator(op)
	if !exists {
		if opt.FailOnUnknownCondition {
			trc.Observation("no match; unknown condition")
			return false, fmt.Errorf("unknown condition operator '%s'", op)
		}
		return true, nil
	}

	// Check condition evaluation against actual values
	for k, v := range cond {
		match := f(ac, trc, k, v)
		if !match {
			trc.Observation("no match; condition evaluated to false")
			return false, nil
		}
	}

	trc.Observation("match!")
	return true, nil
}

// evalPermissionsBoundary assesses the permissions boundary of the Principal to determine whether
// or not it allows the provided AuthContext
func evalPermissionsBoundary(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext) (EffectSet, error) {

	trc.Push("evaluating permission boundaries")
	defer trc.Pop()

	// Empty permissions boundary = allowed; otherwise we have to evaluate
	if ac.Principal.PermissionsBoundary.Empty() {
		effectset := EffectSet{}
		effectset.Add(policy.EFFECT_ALLOW)
		return effectset, nil
	}

	// Specify the statement evaluation funcs we will consider for permission boundary access
	funcs := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesCondition,
	}

	return evalPolicy(trc, opt, ac, ac.Principal.PermissionsBoundary, funcs)
}
