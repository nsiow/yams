package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalPolicy computes whether the provided policy matches the AuthContext
// TODO(nsiow) re-add trace statements to all of the below functions
// (evalPolicy/evalPolicies/evalStatement)
func evalPolicy(s *subject, policy policy.Policy, funcs ...evalFunction) Decision {
	decision := Decision{}

	for _, stmt := range policy.Statement {
		effect := evalStatement(s, stmt, funcs)
		decision.Merge(effect)
	}

	return decision
}
