package sim

import (
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/gate"
	"github.com/nsiow/yams/pkg/sim/wildcard"
)

// evalStatement computes whether the provided statements match the AuthContext
func evalStatement(s *subject, stmt policy.Statement, funcs []evalFunction) Decision {
	decision := Decision{}

	for _, f := range funcs {
		if !f(s, &stmt) {
			return decision
		}
	}

	decision.Add(stmt.Effect)
	return decision
}

// evalStatementMatchesAction computes whether the Statement matches the AuthContext's Action
func evalStatementMatchesAction(s *subject, stmt *policy.Statement) bool {

	s.trc.Push("evaluating Action")
	defer s.trc.Pop()

	// Handle empty Action
	if s.ac.Action == nil {
		s.trc.Observation("AuthContext missing Action")
		return false
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

	shortName := s.ac.Action.ShortName()
	for _, a := range action {
		match := wildcard.MatchSegmentsIgnoreCase(a, shortName)
		if match {
			s.trc.Attr("action", a)
			s.trc.Observation("action matched")
			return _gate.Apply(true)
		}
	}

	s.trc.Observation("no actions match")
	return _gate.Apply(false)
}

// evalStatementMatchesPrincipal computes whether the Statement matches the AuthContext's Principal
func evalStatementMatchesPrincipal(s *subject, stmt *policy.Statement) bool {

	s.trc.Push("evaluating Principal")
	defer s.trc.Pop()

	if s.ac.Principal == nil {
		s.trc.Observation("AuthContext missing Principal")
		return false
	}

	var _gate gate.Gate
	var principals policy.Principal
	switch {
	case stmt.Principal.All:
		s.trc.Observation("saw special Principal=* block")
		return true
	case stmt.NotPrincipal.All:
		s.trc.Observation("saw special NotPrincipal=* block")
		return false
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
			return _gate.Apply(true)
		}
	}

	return _gate.Apply(false)
}

// evalStatementMatchesPrincipalExact computes whether the Statement matches the AuthContext's
// Principal using an exact-match criteria (no wildcards)
func evalStatementMatchesPrincipalExact(s *subject, stmt *policy.Statement) bool {

	s.trc.Push("evaluating Principal exact-match case")
	defer s.trc.Pop()

	if s.ac.Principal == nil {
		s.trc.Observation("AuthContext missing Principal")
		return false
	}

	return stmt.Principal.AWS.Contains(s.ac.Principal.Arn)
}

// evalStatementMatchesResource computes whether the Statement matches the AuthContext's Resource
func evalStatementMatchesResource(s *subject, stmt *policy.Statement) bool {

	s.trc.Push("evaluating Resource")
	defer s.trc.Pop()

	// Handle empty Resource
	if s.ac.Resource == nil {
		s.trc.Observation("AuthContext missing Resource")
		return false
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
			return _gate.Apply(true)
		}
	}

	return _gate.Apply(false)
}

// evalStatementMatchesCondition computes whether the Statement's Conditions hold true given the
// provided AuthContext
func evalStatementMatchesCondition(s *subject, stmt *policy.Statement) bool {

	s.trc.Push("evaluating Condition")
	defer s.trc.Pop()

	for op, cond := range stmt.Condition {
		if !evalCheckCondition(s, op, cond) {
			s.trc.Observation("condition evaluated to false")
			return false
		}
	}

	s.trc.Observation("condition evaluated to true")
	return true
}
