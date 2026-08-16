package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalSCP assesses the service control policies of the Resource to determine whether or not it
// allows the provided AuthContext
func evalSCP(s *subject) Decision {
	trc := s.trc.Enabled()
	if trc {
		s.trc.Push("evaluating service control policies")
		defer s.trc.Pop()
	}

	decision := Decision{}

	// Empty SCP = allowed; otherwise we have to evaluate. Inspect every node, not just
	// the first, so an account-level SCP isn't skipped when the root has none.
	hasAnySCP := false
	for nodeIndex := range s.auth.Principal.Account.OrgNodes {
		node := &s.auth.Principal.Account.OrgNodes[nodeIndex]
		if len(node.SCPs) > 0 {
			hasAnySCP = true
			break
		}
	}
	if len(s.auth.Principal.Account.OrgNodes) == 0 || !hasAnySCP {
		if trc {
			s.trc.Log("skipping SCPs: none found")
		}
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	if principalIsServiceLinkedRole(s) || principalIsInManagementAccount(s) {
		if trc {
			s.trc.Log("skipping SCPs: exempt principal")
		}
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	// Iterate through layers of SCP, only continuing if we get an allow result through each layer
	for nodeIndex := range s.auth.Principal.Account.OrgNodes {
		node := &s.auth.Principal.Account.OrgNodes[nodeIndex]
		if trc {
			s.trc.Push("evaluating service control policies for node: %s of type %s", node.Name, node.Type)
		}
		layerDecision := Decision{}

		for scpIndex := range node.SCPs {
			scp := &node.SCPs[scpIndex]
			if trc {
				s.trc.Push("evaluating service control policy: %s", scp.Name)
			}

			localDecision := evalPolicy(s, &scp.Policy,
				evalStatementMatchesAction,
				evalStatementMatchesPrincipal,
				evalStatementMatchesResource,
				evalStatementMatchesCondition)
			if trc && localDecision.DeniedExplicit() {
				s.trc.Denied("explicit deny in service control policy: %s", scp.Name)
			}

			// Calculate access for this layer
			layerDecision.Merge(localDecision)

			if trc {
				s.trc.Pop()
			}
		}

		if !layerDecision.Allowed() {
			if trc {
				s.trc.Log("deny due to SCPs for node: %s of type %s", node.Name, node.Type)
				s.trc.Pop()
			}
			return layerDecision
		}

		decision.Merge(layerDecision)
		if trc {
			s.trc.Pop()
		}
	}

	return decision
}
