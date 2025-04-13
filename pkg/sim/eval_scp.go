package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// evalSCP assesses the service control policies of the Principal to determine whether or not it
// allows the provided AuthContext
func evalSCP(s *subject) Decision {

	s.trc.Push("evaluating service control policies")
	defer s.trc.Pop()

	// Missing account or empty SCP = allowed; otherwise we have to evaluate
	account := s.ac.Principal.Account
	if len(account.SCPs) == 0 {
		decision := Decision{}
		decision.Add(policy.EFFECT_ALLOW)
		return decision
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
		layerDecision = evalControlPolicyLayer(s, account, funcs, layer, i+1)

		// If not allowed at this layer, propagate result up; should be denied
		if !layerDecision.Allowed() {
			s.trc.Observation(
				fmt.Sprintf("SCP access denied at layer %d of %d", i+1, len(account.SCPs)))
			return layerDecision
		}
	}

	return layerDecision
}

// evalControlPolicyLayer evaluates a single "layer" of control policies
//
// This is separated since each logical layer must result in an ALLOW decision in order to
// continue
func evalControlPolicyLayer(
	s *subject,
	account entities.Account,
	funcs []evalFunction,
	layer []policy.Policy,
	layerId int) Decision {

	s.trc.Push(fmt.Sprintf("evaluating SCP layer %d of %d", layerId, len(account.SCPs)))
	defer s.trc.Pop()

	decision := Decision{}
	for _, pol := range layer {
		effect := evalPolicy(s, pol, funcs)
		decision.Merge(effect)
	}

	return decision
}
