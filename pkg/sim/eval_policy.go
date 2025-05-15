package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalPolicy computes whether the provided policy matches the AuthContext
// TODO(nsiow) re-add trace statements to all of the below functions
// (evalPolicy/evalPolicies/evalStatement)
func evalPolicy(s *subject, policy policy.Policy, funcs ...evalFunction) Decision {
	s.trc.Push("evaluating policy: %s", Id(policy.Id, 0))
	defer s.trc.Pop()

	decision := Decision{}

	for i, stmt := range policy.Statement {
		s.trc.Push("evaluating statement: %s", Id(stmt.Sid, i))

		effect := evalStatement(s, stmt, funcs)
		decision.Merge(effect)

		s.trc.Pop()
	}

	return decision
}
