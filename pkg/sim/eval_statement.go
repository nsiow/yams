package sim

import (
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/gate"
	"github.com/nsiow/yams/pkg/sim/trace"
	"github.com/nsiow/yams/pkg/sim/wildcard"
)

// evalStatement computes whether the provided statements match the AuthContext
func evalStatement(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	stmt policy.Statement,
	funcs []evalFunction) (Decision, error) {

	for _, f := range funcs {
		match, err := f(trc, opt, ac, &stmt)
		if err != nil {
			return Decision{}, err
		}

		if !match {
			return Decision{}, nil
		}
	}

	decision := Decision{}
	decision.Add(stmt.Effect)
	return decision, nil
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
