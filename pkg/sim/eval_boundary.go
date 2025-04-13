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
	if s.ac.Principal.PermissionsBoundary.Empty() {
		decision := Decision{}
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	// Specify the statement evaluation funcs we will consider for permission boundary access
	funcs := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesCondition,
	}

	return evalPolicy(s, s.ac.Principal.PermissionsBoundary, funcs)
}
