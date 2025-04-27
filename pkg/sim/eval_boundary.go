package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalPermissionsBoundary assesses the permissions boundary of the Principal to determine whether
// or not it allows the provided AuthContext
func evalPermissionsBoundary(s *subject) Decision {

	s.trc.Push("evaluating permission boundaries")
	defer s.trc.Pop()

	// Empty permissions boundary = allowed; otherwise we have to evaluate
	boundary := s.auth.Principal.PermissionBoundary.Policy
	if boundary.Empty() {
		s.trc.Log("no permission boundary found")
		decision := Decision{}
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	s.trc.Push("evaluating permission boundary: %s", s.auth.Principal.PermissionBoundary.Arn)
	defer s.trc.Pop()

	return evalPolicy(s, boundary,
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesCondition,
	)
}
