package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalPolicy computes whether the provided policy matches the AuthContext
func evalPolicy(s *subject, policy policy.Policy, funcs ...evalFunction) Decision {
	trc := s.trc.Enabled()
	if trc {
		s.trc.Push("evaluating policy: %s", Id(policy.Id, 0))
		defer s.trc.Pop()
	}

	decision := Decision{}

	for i, stmt := range policy.Statement {
		if trc {
			s.trc.Push("evaluating statement: %s", Id(stmt.Sid, i))
		}

		effect := evalStatement(s, stmt, funcs)
		decision.Merge(effect)

		if trc {
			s.trc.Pop()
		}
	}

	return decision
}
