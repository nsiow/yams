package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalPolicies computes whether the provided policies match the AuthContext
func evalPolicies(s *subject, policies []policy.Policy, funcs []evalFunction) Decision {
	decision := Decision{}

	for _, pol := range policies {
		effect := evalPolicy(s, pol, funcs)
		decision.Merge(effect)
	}

	return decision
}

// evalPolicy computes whether the provided policy matches the AuthContext
// TODO(nsiow) re-add trace statements to all of the below functions
// (evalPolicy/evalPolicies/evalStatement)
func evalPolicy(s *subject, policy policy.Policy, funcs []evalFunction) Decision {
	decision := Decision{}

	for _, stmt := range policy.Statement {
		effect := evalStatement(s, stmt, funcs)
		decision.Merge(effect)
	}

	return decision
}
