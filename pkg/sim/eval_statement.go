package sim

import (
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/gate"
	"github.com/nsiow/yams/pkg/sim/wildcard"
)

// evalStatement computes whether the provided statements match the AuthContext
func evalStatement(s *subject, stmt policy.Statement, funcs []evalFunction) (Decision, error) {

	for _, f := range funcs {
		match, err := f(s, &stmt)
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
func evalStatementMatchesAction(s *subject, stmt *policy.Statement) (bool, error) {

	s.trc.Push("evaluating Action")
	defer s.trc.Pop()

	// Handle empty Action
	if s.ac.Action == nil {
		s.trc.Observation("AuthContext missing Action")
		return false, nil
	}

	// Determine which Action block to use
	var _gate gate.Gate
	var action policy.Action
	if !stmt.Action.Empty() {
		s.trc.Observation("using Action block")
		action = stmt.Action
	} else {
		s.trc.Observation("using NotAction block")
		action = stmt.NotAction
		_gate.Invert()
	}

	for _, a := range action {
		match := wildcard.MatchSegmentsIgnoreCase(a, s.ac.Action.Name)
		if match {
			s.trc.Attr("action", a)
			s.trc.Observation("action matched")
			return _gate.Apply(true), nil
		}
	}

	s.trc.Observation("no actions match")
	return _gate.Apply(false), nil
}

// evalStatementMatchesPrincipal computes whether the Statement matches the AuthContext's Principal
func evalStatementMatchesPrincipal(s *subject, stmt *policy.Statement) (bool, error) {

	s.trc.Push("evaluating Principal")
	defer s.trc.Pop()

	// Handle empty Principal
	if s.ac.Principal == nil {
		s.trc.Observation("AuthContext missing Principal")
		return false, nil
	}

	// Determine which Principal block to use
	var _gate gate.Gate
	var principals policy.Principal
	switch {
	case stmt.Principal.All:
		s.trc.Observation("saw special Principal=* block")
		return true, nil
	case stmt.NotPrincipal.All:
		s.trc.Observation("saw special NotPrincipal=* block")
		return false, nil
	case !stmt.Principal.Empty():
		s.trc.Observation("using Principal block")
		principals = stmt.Principal
	default:
		s.trc.Observation("using NotPrincipal block")
		principals = stmt.NotPrincipal
		_gate.Invert()
	}

	// TODO(nsiow) validate that this is how Principals are evaluated - exact matches?
	for _, p := range principals.AWS {
		match := wildcard.MatchAllOrNothing(p, s.ac.Principal.Arn)
		if match {
			return _gate.Apply(true), nil
		}
	}

	return _gate.Apply(false), nil
}

// evalStatementMatchesResource computes whether the Statement matches the AuthContext's Resource
func evalStatementMatchesResource(s *subject, stmt *policy.Statement) (bool, error) {

	s.trc.Push("evaluating Resource")
	defer s.trc.Pop()

	// Handle empty Resource
	if s.ac.Resource == nil {
		s.trc.Observation("AuthContext missing Resource")
		return false, nil
	}

	// Determine which Resource block to use
	var _gate gate.Gate
	var resources policy.Resource
	if !stmt.Resource.Empty() {
		s.trc.Observation("using Resource block")
		resources = stmt.Resource
	} else {
		s.trc.Observation("using NotResource block")
		resources = stmt.NotResource
		_gate.Invert()
	}

	// TODO(nsiow) this may need to change for subresource based operations e.g. s3:getobject
	// TODO(nsiow) this needs to support variable expansion
	for _, r := range resources {
		match := wildcard.MatchSegments(r, s.ac.Resource.Arn)
		if match {
			return _gate.Apply(true), nil
		}
	}

	return _gate.Apply(false), nil
}

// evalStatementMatchesCondition computes whether the Statement's Conditions hold true given the
// provided AuthContext
func evalStatementMatchesCondition(s *subject, stmt *policy.Statement) (bool, error) {

	s.trc.Push("evaluating Condition")
	defer s.trc.Pop()

	for op, cond := range stmt.Condition {
		result, err := evalCheckCondition(s, op, cond)
		if err != nil {
			return false, err
		}

		if !result {
			s.trc.Observation("condition evaluated to false")
			return false, nil
		}
	}

	s.trc.Observation("condition evaluated to true")
	return true, nil
}
