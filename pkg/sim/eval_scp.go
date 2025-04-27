package sim

import (
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// evalSCP assesses the service control policies of the Principal to determine whether or not it
// allows the provided AuthContext
func evalSCP(s *subject) Decision {
	s.trc.Push("evaluating service control policies")
	defer s.trc.Pop()

	decision := Decision{}

	// Empty SCP = allowed; otherwise we have to evaluate
	if len(s.auth.Principal.Account.SCPs) == 0 {
		s.trc.Log("no SCPs found")
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	// Iterate through layers of SCP, only continuing if we get an allow result through each layer
	scps := s.auth.Principal.Account.SCPs
	for i, layer := range scps {

		s.trc.Push("evaluating SCP layer %d of %d", i, len(scps)-1)

		// Calculate access for this layer
		decision = evalSCPLayer(s, layer)

		// If not allowed at this layer, propagate result up; should be denied
		if !decision.Allowed() {
			s.trc.Pop()
			return decision
		}

		s.trc.Pop()
	}

	return decision
}

// evalSCPLayer evaluates a single "layer" of service control policies
//
// This is separated since each logical layer must result in an ALLOW decision in order to
// continue
func evalSCPLayer(s *subject, layer []entities.ManagedPolicy) (decision Decision) {
	for _, pol := range layer {
		s.trc.Push("evaluating SCP: %s", pol.Arn)
		decision.Merge(
			evalPolicy(s, pol.Policy,
				evalStatementMatchesAction,
				evalStatementMatchesResource,
				evalStatementMatchesCondition,
			),
		)
		s.trc.Pop()
	}

	return decision
}
