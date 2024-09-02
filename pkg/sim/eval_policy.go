package sim

import (
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/trace"
)

// evalPolicies computes whether the provided policies match the AuthContext
func evalPolicies(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	policies []policy.Policy,
	funcs []evalFunction) (Decision, error) {

	decision := Decision{}

	for _, pol := range policies {
		eff, err := evalPolicy(trc, opt, ac, pol, funcs)
		if err != nil {
			return eff, err
		}

		// TODO(nsiow) short circuit on Deny, or keep going for completeness?

		decision.Merge(eff)
	}

	return decision, nil
}

// evalPolicy computes whether the provided policy matches the AuthContext
// FIXME(nsiow) re-add trace statements to all of the below functions
// (evalPolicy/evalPolicies/evalStatement)
func evalPolicy(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	policy policy.Policy,
	funcs []evalFunction) (Decision, error) {

	decision := Decision{}

	for _, stmt := range policy.Statement {
		eff, err := evalStatement(trc, opt, ac, stmt, funcs)
		if err != nil {
			return eff, err
		}

		decision.Merge(eff)
	}

	return decision, nil
}
