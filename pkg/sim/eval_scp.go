package sim

import (
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// evalSCP assesses the service control policies of the Principal to determine whether or not it
// allows the provided AuthContext
func evalSCP(s *subject) (decision Decision) {
	s.trc.Push("evaluating service control policies")
	defer s.trc.Pop()

	// Missing account or empty SCP = allowed; otherwise we have to evaluate
	account := s.ac.Principal.FrozenAccount
	if len(account.FrozenSCPs) == 0 {
		// TODO(nsiow) add observation for missing SCPs
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	// Iterate through layers of SCP, only continuing if we get an allow result through each layer
	// TODO(nsiow) add better tracing here
	for i, layer := range account.FrozenSCPs {

		// Calculate access for this layer
		decision = evalControlPolicyLayer(s, layer)

		// If not allowed at this layer, propagate result up; should be denied
		if !decision.Allowed() {
			s.trc.Observation("SCP access denied at layer %d of %d", i, len(account.FrozenSCPs)-1)
			return decision
		}
	}

	return decision
}

// evalControlPolicyLayer evaluates a single "layer" of control policies
//
// This is separated since each logical layer must result in an ALLOW decision in order to
// continue
func evalControlPolicyLayer(s *subject, layer []entities.ManagedPolicy) (decision Decision) {
	for _, pol := range layer {
		decision.Merge(
			evalPolicy(s, pol.Policy,
				evalStatementMatchesAction,
				evalStatementMatchesResource,
				evalStatementMatchesCondition,
			),
		)
	}

	return decision
}
