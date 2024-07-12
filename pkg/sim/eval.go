package sim

import (
	"errors"
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	es "github.com/nsiow/yams/pkg/sim/effectset"
	"github.com/nsiow/yams/pkg/sim/gate"
	"github.com/nsiow/yams/pkg/sim/trace"
	"github.com/nsiow/yams/pkg/sim/wildcard"
)

// evalOverallAccess calculates both Principal + Resource access same performs both same-account
// and different-account evaluations
func evalOverallAccess(opts *Options, ac AuthContext) (*Result, error) {

	trc := trace.New()
	res := Result{Trace: trc}

	trc.Attr("authContext", ac)

	// Calculate Principal access
	pAccess, err := evalPrincipalAccess(opts, ac, trc)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error evaluating principal access"), err)
	}
	// ... check for explicit Deny results
	if pAccess.Contains(policy.EFFECT_DENY) {
		res.IsAllowed = false
		trc.Decision("[explicit deny] found in identity policy")
		return &res, nil
	}

	// Calculate Resource access
	rAccess, err := evalResourceAccess(opts, ac, trc)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error evaluating resource access"), err)
	}
	// ... check for explicit Deny results
	if rAccess.Contains(policy.EFFECT_DENY) {
		res.IsAllowed = false
		trc.Decision("[explicit deny] found in resource policy")
		return &res, nil
	}

	// If same account, access is granted if the Principal has access
	if evalIsSameAccount(ac.Principal, ac.Resource) {
		if pAccess.Contains(policy.EFFECT_ALLOW) {
			res.IsAllowed = true
			trc.Decision("[allow] access granted via same-account identity policy")
			return &res, nil
		}

		// TODO(nsiow) implement correct behavior for same-account access via explicit ARN
		res.IsAllowed = false
		trc.Decision("[implicit deny] no identity-based policy allows this action")
		return &res, nil
	}

	// If x-account, access is granted if the Principal has access and the Resource permits that
	// access
	if pAccess.Contains(policy.EFFECT_ALLOW) && rAccess.Contains(policy.EFFECT_ALLOW) {
		res.IsAllowed = true
		trc.Decision("[allow] access granted via x-account identity + resource policies")
		return &res, nil
	}
	if pAccess.Contains(policy.EFFECT_ALLOW) && !rAccess.Contains(policy.EFFECT_ALLOW) {
		res.IsAllowed = false
		trc.Decision("[implicit deny] x-account, missing resource policy access")
		return &res, nil
	}
	if !pAccess.Contains(policy.EFFECT_ALLOW) && rAccess.Contains(policy.EFFECT_ALLOW) {
		res.IsAllowed = false
		trc.Decision("[implicit deny] x-account, missing identity policy access")
		return &res, nil
	}

	// We fell through and no access was granted from either side
	res.IsAllowed = false
	trc.Decision("[implicit deny] x-account, missing both identity + resource access")
	return &res, nil
}

// evalFunction is the blueprint of a function that allows us to evaluate a single statement
type evalFunction func(*Options, AuthContext, *trace.Trace, *policy.Statement) (bool, error)

// evalPrincipalAccess calculates the Principal-side access to the specified Resource
func evalPrincipalAccess(opts *Options, ac AuthContext, trc *trace.Trace) (*es.EffectSet, error) {

	// Create new trace frame
	trc.Push("evaluating principal policies")
	defer trc.Pop()

	// Specify the types of policies we will consider for Principal access
	effectivePolicies := map[string][]policy.Policy{
		"inline":  ac.Principal.InlinePolicies,
		"managed": ac.Principal.AttachedPolicies,
		"group":   ac.Principal.GroupPolicies,
	}

	// Specify the statement evaluation functions we will consider for Principal access
	functions := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesCondition,
	}

	// Iterate over policy types / policies / statements to evaluate access
	effects := es.EffectSet{}
	for policytype, policies := range effectivePolicies {
		trc.Push(fmt.Sprintf("policytype=%s", policytype))
		for i, pol := range policies {
			trc.Push(fmt.Sprintf("policy=%s", Id(i, pol.Id)))
			for j, stmt := range pol.Statement {
				trc.Push(fmt.Sprintf("stmt=%s", Id(j, stmt.Sid)))

				matchedAll := true
				for _, f := range functions {
					match, err := f(opts, ac, trc, &stmt)
					if err != nil {
						return nil, errors.Join(
							fmt.Errorf("error evaluating principal policy statement[sid=%s]", stmt.Sid),
							err)
					}
					if !match {
						matchedAll = false
						break
					}
				}

				if matchedAll {
					effects.Add(stmt.Effect)
					trc.Attr("effect", stmt.Effect)
					trc.Decision("statement matches, adding Effect")
				} else {
					trc.Decision("statement does not match")
				}
				trc.Pop()
			}
			trc.Pop()
		}
		trc.Pop()
	}

	return &effects, nil
}

// evalResourceAccess calculates the Resource-side access with regard to the specified Principal
func evalResourceAccess(opts *Options, ac AuthContext, trc *trace.Trace) (*es.EffectSet, error) {

	// Create new trace frame
	trc.Push("evaluating resource policies")
	defer trc.Pop()

	// Specify the statement evaluation functions we will consider for Principal access
	functions := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesPrincipal,
		evalStatementMatchesCondition,
	}

	// Iterate over resource policy statements to evaluate access
	effects := es.EffectSet{}
	for i, stmt := range ac.Resource.Policy.Statement {
		trc.Push(fmt.Sprintf("stmt=%s", Id(i, stmt.Sid)))
		matchedAll := true
		for _, f := range functions {
			match, err := f(opts, ac, trc, &stmt)
			if err != nil {
				return nil, errors.Join(
					fmt.Errorf("error evaluating principal policy statement[sid=%s]", stmt.Sid),
					err)
			}
			if !match {
				matchedAll = false
				break
			}
		}

		if matchedAll {
			effects.Add(stmt.Effect)
		}
		trc.Pop()
	}

	return &effects, nil

}

// evalStatementMatchesAction computes whether the Statement matches the AuthContext's Action
func evalStatementMatchesAction(
	opts *Options, ac AuthContext, trc *trace.Trace, stmt *policy.Statement) (bool, error) {

	trc.Push("evaluating Action")
	defer trc.Pop()

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
	opts *Options, ac AuthContext, trc *trace.Trace, stmt *policy.Statement) (bool, error) {

	trc.Push("evaluating Principal")
	defer trc.Pop()

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
	opts *Options, ac AuthContext, trc *trace.Trace, stmt *policy.Statement) (bool, error) {

	trc.Push("evaluating Resource")
	defer trc.Pop()

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
	opts *Options, ac AuthContext, trc *trace.Trace, stmt *policy.Statement) (bool, error) {

	trc.Push("evaluating Condition")
	defer trc.Pop()

	for op, cond := range stmt.Condition {
		result, err := evalCheckCondition(opts, ac, trc, op, cond)
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
	opts *Options, ac AuthContext, trc *trace.Trace,
	op string, cond map[string]policy.Value) (bool, error) {

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
		if opts.FailOnUnknownCondition {
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

// evalIsSameAccount determines whether or not the provided Principal + Resource exist within the
// same AWS account
func evalIsSameAccount(p *entities.Principal, r *entities.Resource) bool {
	return p.Account == r.Account
}
