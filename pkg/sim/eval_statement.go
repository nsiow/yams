package sim

import (
	"fmt"
	"slices"

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
	if s.auth.Action == nil {
		s.trc.Log("AuthContext missing Action")
		return false
	}

	// Determine which Action block to use
	var _gate gate.Gate
	var action policy.Action
	if !stmt.Action.Empty() {
		s.trc.Log("using Action block")
		action = stmt.Action
	} else {
		s.trc.Log("using NotAction block")
		action = stmt.NotAction
		_gate.Invert()
	}

	shortName := s.auth.Action.ShortName()
	for _, a := range action {
		match := wildcard.MatchSegmentsIgnoreCase(a, shortName)
		if match {
			s.trc.Log("match: %s and %s", a, shortName)
			return _gate.Apply(true)
		}
	}

	s.trc.Log("action does not match")
	return _gate.Apply(false)
}

// evalStatementMatchesPrincipal computes whether the Statement matches the AuthContext's Principal
func evalStatementMatchesPrincipal(s *subject, stmt *policy.Statement) bool {

	s.trc.Push("evaluating Principal")
	defer s.trc.Pop()

	if s.auth.Principal == nil {
		s.trc.Log("AuthContext missing Principal")
		return false
	}

	var _gate gate.Gate
	var principals policy.Principal
	switch {
	case stmt.Principal.All:
		s.trc.Log("saw special Principal=* block")
		return true
	case stmt.NotPrincipal.All:
		s.trc.Log("saw special NotPrincipal=* block")
		return false
	case !stmt.Principal.Empty():
		s.trc.Log("using Principal block")
		principals = stmt.Principal
	default:
		s.trc.Log("using NotPrincipal block")
		principals = stmt.NotPrincipal
		_gate.Invert()
	}

	for _, p := range principals.AWS {
		// Handle account-root syntax
		if isAccountRootMatch(p, s.auth.Principal.AccountId) ||
			wildcard.MatchAllOrNothing(p, s.auth.Principal.Arn) {
			s.trc.Log("match: %s and %s", p, s.auth.Principal.Arn)
			return _gate.Apply(true)
		}
	}

	s.trc.Log("principal does not match")
	return _gate.Apply(false)
}

// evalStatementMatchesPrincipalExact computes whether the Statement matches the AuthContext's
// Principal using an exact-match criteria (no wildcards)
func evalStatementMatchesPrincipalExact(s *subject, stmt *policy.Statement) bool {

	s.trc.Push("evaluating Principal exact-match case")
	defer s.trc.Pop()

	if s.auth.Principal == nil {
		s.trc.Log("AuthContext missing Principal")
		return false
	}

	result := stmt.Principal.AWS.Contains(s.auth.Principal.Arn)
	s.trc.Log("result: %v", result)
	return result
}

// evalStatementMatchesResource computes whether the Statement matches the AuthContext's Resource
func evalStatementMatchesResource(s *subject, stmt *policy.Statement) bool {

	s.trc.Push("evaluating Resource")
	defer s.trc.Pop()

	// Handle empty Resource
	if s.auth.Resource == nil && s.auth.Action.HasTargets() {
		s.trc.Log("AuthContext missing Resource")
		return false
	}

	// Determine which Resource block to use
	var _gate gate.Gate
	var resources policy.Resource
	if !stmt.Resource.Empty() {
		s.trc.Log("using Resource block")
		resources = stmt.Resource
	} else {
		s.trc.Log("using NotResource block")
		resources = stmt.NotResource
		_gate.Invert()
	}

	// Handle the case of resource-less API calls (Lists, Describes, etc)
	if !s.auth.Action.HasTargets() {
		return _gate.Apply(slices.Contains(resources, "*"))
	}

	// TODO(nsiow) this may need to change for subresource based operations e.g. s3:getobject
	// TODO(nsiow) this needs to support variable expansion
	for _, r := range resources {
		match := wildcard.MatchSegments(r, s.auth.Resource.Arn)
		if match {
			s.trc.Log("match: %s and %s", r, s.auth.Resource.Arn)
			return _gate.Apply(true)
		}
	}

	s.trc.Log("resource does not match")
	return _gate.Apply(false)
}

// evalStatementMatchesCondition computes whether the Statement's Conditions hold true given the
// provided AuthContext
func evalStatementMatchesCondition(s *subject, stmt *policy.Statement) bool {
	if len(stmt.Condition) == 0 {
		s.trc.Log("skipping condition: none found")
		return true
	}

	s.trc.Push("evaluating Condition")
	defer s.trc.Pop()

	for op, cond := range stmt.Condition {
		if !evalCheckCondition(s, op, cond) {
			s.trc.Log("condition evaluated to false")
			return false
		}
	}

	s.trc.Log("condition evaluated to true")
	return true
}

// isAccountRootMatch handles the unique delegation for account roots in IAM policies
func isAccountRootMatch(pattern string, principalAccountId string) bool {
	return pattern == principalAccountId ||
		pattern == fmt.Sprintf("arn:aws:iam::%s:root", principalAccountId)
}
