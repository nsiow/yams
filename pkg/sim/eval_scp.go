package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalSCP assesses the service control policies of the Resource to determine whether or not it
// allows the provided AuthContext
func evalSCP(s *subject) Decision {
	s.trc.Push("evaluating service control policies")
	defer s.trc.Pop()

	decision := Decision{}

	// Empty SCP = allowed; otherwise we have to evaluate
	if len(s.auth.Principal.Account.OrgNodes) == 0 ||
		len(s.auth.Principal.Account.OrgNodes[0].SCPs) == 0 {
		s.trc.Log("no SCPs found")
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	// Iterate through layers of SCP, only continuing if we get an allow result through each layer
	for _, node := range s.auth.Principal.Account.OrgNodes {

		s.trc.Push("evaluating service control policies for node: %s of type %s", node.Name, node.Type)
		layerDecision := Decision{}

		for i, scp := range node.SCPs {
			s.trc.Push("evaluating service control policy: %s", Id(scp.Policy.Id, i))

			// Calculate access for this layer
			layerDecision.Merge(
				evalPolicy(s, scp.Policy,
					evalStatementMatchesAction,
					evalStatementMatchesResource,
					evalStatementMatchesCondition,
				),
			)
			s.trc.Pop()
		}

		if !layerDecision.Allowed() {
			s.trc.Log("deny due to SCPs for node: %s of type %s", node.Name, node.Type)
			s.trc.Pop()
			return layerDecision
		}

		decision.Merge(layerDecision)
		s.trc.Pop()
	}

	return decision
}
