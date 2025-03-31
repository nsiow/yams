package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/trace"
)

// evalSCP assesses the service control policies of the Principal to determine whether or not it
// allows the provided AuthContext
func evalSCP(trc *trace.Trace, opt *Options, ac AuthContext) (Decision, error) {

	trc.Push("evaluating service control policies")
	defer trc.Pop()

	// Missing account or empty SCP = allowed; otherwise we have to evaluate
	account := ac.Principal.Account
	if len(account.SCPs) == 0 {
		decision := Decision{}
		decision.Add(policy.EFFECT_ALLOW)
		return decision, nil
	}

	// Specify the statement evaluation funcs we will consider for SCP access
	funcs := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesCondition,
	}

	layerDecision := Decision{}

	// Iterate through layers of SCP, only continuing if we get an allow result through each layer
	for i, layer := range account.SCPs {

		// Calculate access for this layer
		var err error
		layerDecision, err = evalControlPolicyLayer(trc, opt, ac, account, funcs, layer, i+1)
		if err != nil {
			return layerDecision, err
		}

		// If not allowed at this layer, propagate result up; should be denied
		if !layerDecision.Allowed() {
			trc.Observation(
				fmt.Sprintf("SCP access denied at layer %d of %d", i+1, len(account.SCPs)))
			return layerDecision, nil
		}
	}

	return layerDecision, nil
}

// evalControlPolicyLayer evaluates a single "layer" of control policies
//
// This is separated since each logical layer must result in an ALLOW decision in order to
// continue
func evalControlPolicyLayer(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	account entities.Account,
	funcs []evalFunction,
	layer []policy.Policy,
	layerId int) (Decision, error) {

	trc.Push(fmt.Sprintf("evaluating SCP layer %d of %d", layerId, len(account.SCPs)))
	defer trc.Pop()

	decision := Decision{}
	for _, pol := range layer {
		result, err := evalPolicy(trc, opt, ac, pol, funcs)
		if err != nil {
			return decision, err
		}

		// TODO(nsiow) determine performance vs insight tradeoff of short-circuiting here
		// if result.ExplicitlyDenied() {
		// 	return decision, nil
		// }

		decision.Merge(result)
	}

	return decision, nil
}
