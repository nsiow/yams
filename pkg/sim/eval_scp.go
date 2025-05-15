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
		s.trc.Log("skipping SCPs: none found")
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	// Iterate through layers of SCP, only continuing if we get an allow result through each layer
	for _, node := range s.auth.Principal.Account.OrgNodes {

		s.trc.Push("evaluating service control policies for node: %s of type %s", node.Name, node.Type)
		layerDecision := Decision{}

		for _, scp := range node.SCPs {
			// FIXME(nsiow) figure out how to use scp.Name instead and do the same for RCP
			s.trc.Push("evaluating service control policy: %s", scp.Arn)

			localDecision := evalPolicy(s, scp.Policy,
				evalStatementMatchesAction,
				evalStatementMatchesResource,
				evalStatementMatchesCondition)
			if localDecision.DeniedExplicit() {
				s.trc.Denied("explicit deny in service control policy: %s", scp.Arn)
			}

			// Calculate access for this layer
			layerDecision.Merge(localDecision)

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
