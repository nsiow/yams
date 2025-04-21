package sim

import (
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// evalRCP assesses the resource control policies of the Resource to determine whether or not it
// allows the provided AuthContext
func evalRCP(s *subject) Decision {
	s.trc.Push("evaluating resource control policies")
	defer s.trc.Pop()

	decision := Decision{}

	// Missing resource or empty RCP = allowed; otherwise we have to evaluate
	if s.auth.Resource == nil || len(s.auth.Resource.Account.RCPs) == 0 {
		// TODO(nsiow) add observation for missing SCPs
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	// Iterate through layers of RCP, only continuing if we get an allow result through each layer
	// TODO(nsiow) add better tracing here
	rcps := s.auth.Resource.Account.RCPs
	for i, layer := range rcps {

		// Calculate access for this layer
		decision = evalRCPLayer(s, layer)

		// If not allowed at this layer, propagate result up; should be denied
		if !decision.Allowed() {
			s.trc.Observation("RCP access denied at layer %d of %d", i, len(rcps)-1)
			return decision
		}
	}

	return decision
}

// evalRCPLayer evaluates a single "layer" of resource control policies
//
// This is separated since each logical layer must result in an ALLOW decision in order to
// continue
func evalRCPLayer(s *subject, layer []entities.ManagedPolicy) (decision Decision) {
	for _, pol := range layer {
		decision.Merge(
			evalPolicy(s, pol.Policy,
				evalStatementMatchesAction,
				evalStatementMatchesPrincipal,
				evalStatementMatchesResource,
				evalStatementMatchesCondition,
			),
		)
	}

	return decision
}
