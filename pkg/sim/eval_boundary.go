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
	boundary := s.ac.Principal.FrozenPermissionBoundary.Policy
	if boundary.Empty() {
		decision := Decision{}
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	return evalPolicy(s, boundary,
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesCondition,
	)
}
