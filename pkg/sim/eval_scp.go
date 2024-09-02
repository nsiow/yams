package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/trace"
)

// evalSCP assesses the service control policies of the Principal to determine whether or not it
// allows the provided AuthContext
func evalSCP(trc *trace.Trace, opt *Options, ac AuthContext) (EffectSet, error) {

	trc.Push("evaluating service control policies")
	defer trc.Pop()

	// Empty SCP = allowed; otherwise we have to evaluate
	if len(ac.Principal.SCPs) == 0 {
		effectset := EffectSet{}
		effectset.Add(policy.EFFECT_ALLOW)
		return effectset, nil
	}

	// Specify the statement evaluation funcs we will consider for SCP access
	funcs := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesCondition,
	}

	layerAccess := EffectSet{}

	// Iterate through layers of SCP, only continuing if we get an allow result through each layer
	for i, layer := range ac.Principal.SCPs {

		// Calculate access for this layer
		var err error
		layerAccess, err = evalControlPolicyLayer(trc, opt, ac, funcs, layer, i+1)
		if err != nil {
			return layerAccess, err
		}

		// If not allowed at this layer, propagate result up; should be denied
		if !layerAccess.Allowed() {
			trc.Observation(
				fmt.Sprintf("SCP access denied at layer %d of %d", i+1, len(ac.Principal.SCPs)))
			return layerAccess, nil
		}
	}

	return layerAccess, nil
}

// evalControlPolicyLayer evaluates a single "layer" of control policies
//
// This is separated since each logical layer must result in an ALLOW decision in order to
// continue
func evalControlPolicyLayer(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	funcs []evalFunction,
	layer []policy.Policy,
	layerId int) (EffectSet, error) {

	trc.Push(fmt.Sprintf("evaluating SCP layer %d of %d", layerId, len(ac.Principal.SCPs)))
	defer trc.Pop()

	effectset := EffectSet{}
	for _, pol := range layer {
		result, err := evalPolicy(trc, opt, ac, pol, funcs)
		if err != nil {
			return effectset, err
		}

		// TODO(nsiow) determine performance vs insight tradeoff of short-circuiting here
		// if result.ExplicitlyDenied() {
		// 	return effectset, nil
		// }

		effectset.Merge(result)
	}

	return effectset, nil
}
