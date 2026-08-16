package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalPolicy computes whether the provided policy matches the AuthContext
func evalPolicy(s *subject, pol *policy.Policy, funcs ...evalFunction) Decision {
	trc := s.trc.Enabled()
	if trc {
		s.trc.Push("evaluating policy: %s", Id(pol.Id, 0))
		defer s.trc.Pop()
	}

	previousVersion := s.policyVersion
	s.policyVersion = pol.Version
	if s.policyVersion == "" {
		s.policyVersion = "2008-10-17"
	}
	defer func() {
		s.policyVersion = previousVersion
	}()

	decision := Decision{}

	for i := range pol.Statement {
		stmt := &pol.Statement[i]
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
