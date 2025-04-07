package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalPolicies computes whether the provided policies match the AuthContext
func evalPolicies(s *subject, policies []policy.Policy, funcs []evalFunction) (Decision, error) {

	decision := Decision{}

	for _, pol := range policies {
		eff, err := evalPolicy(s, pol, funcs)
		if err != nil {
			return eff, err
		}

		// TODO(nsiow) short circuit on Deny, or keep going for completeness?

		decision.Merge(eff)
	}

	return decision, nil
}

// evalPolicy computes whether the provided policy matches the AuthContext
// TODO(nsiow) re-add trace statements to all of the below functions
// (evalPolicy/evalPolicies/evalStatement)
func evalPolicy(s *subject, policy policy.Policy, funcs []evalFunction) (Decision, error) {

	decision := Decision{}

	for _, stmt := range policy.Statement {
		effect, err := evalStatement(s, stmt, funcs)
		if err != nil {
			return effect, err
		}

		decision.Merge(effect)
	}

	return decision, nil
}
