package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/trace"
)

// evalPrincipalAccess calculates the Principal-side access to the specified Resource
func evalPrincipalAccess(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext) (EffectSet, error) {

	trc.Push("evaluating principal policies")
	defer trc.Pop()

	// Specify the types of policies we will consider for Principal access
	effectivePolicies := map[string][]policy.Policy{
		"inline":  ac.Principal.InlinePolicies,
		"managed": ac.Principal.AttachedPolicies,
		"group":   ac.Principal.GroupPolicies,
	}

	// Specify the statement evaluation funcs we will consider for Principal access
	funcs := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesCondition,
	}

	// Iterate over policy types / policies / statements to evaluate access
	effects := EffectSet{}
	for policytype, policies := range effectivePolicies {
		trc.Push(fmt.Sprintf("policytype=%s", policytype))
		eff, err := evalPolicies(trc, opt, ac, policies, funcs)
		if err != nil {
			trc.Pop()
			return effects, err
		}

		effects.Merge(eff)
		trc.Pop()
	}

	return effects, nil
}
