package sim

import (
	"strings"

	"github.com/nsiow/yams/pkg/policy"
)

// supportsRCPs determines whether or not the provided auth context has support for RCPs based on:
// https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_rcps.html#rcp-supported-services
func supportsRCPs(s *subject) bool {
	switch s.auth.Resource.Type {
	case
		"AWS::S3::Bucket",
		"AWS::S3::Object",
		"AWS::SQS::Queue",
		"AWS::KMS::Key":
		return true // support RCPs for all operations
	case "AWS::IAM::Role":
		return strings.EqualFold(s.auth.Action.ShortName(), "sts:assumerole") // depends on the API call
	}

	return false
}

// evalRCP assesses the resource control policies of the Resource to determine whether or not it
// allows the provided AuthContext
func evalRCP(s *subject) Decision {
	s.trc.Push("evaluating resource control policies")
	defer s.trc.Pop()

	decision := Decision{}

	// Missing resource or empty RCP = allowed; otherwise we have to evaluate
	if s.auth.Resource == nil ||
		len(s.auth.Resource.Account.OrgNodes) == 0 ||
		len(s.auth.Resource.Account.OrgNodes[0].RCPs) == 0 {
		s.trc.Log("no RCPs found")
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	// If service does not support RCPs, always allowed
	if !supportsRCPs(s) {
		s.trc.Log("action/resource does not support RCPs")
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	// Iterate through layers of RCP, only continuing if we get an allow result through each layer
	for _, node := range s.auth.Resource.Account.OrgNodes {

		s.trc.Push("evaluating resource control policies for node: %s of type %s", node.Name, node.Type)
		layerDecision := Decision{}

		for i, rcp := range node.RCPs {
			s.trc.Push("evaluating resource control policy: %s", Id(rcp.Policy.Id, i))

			// Calculate access for this layer
			layerDecision.Merge(
				evalPolicy(s, rcp.Policy,
					evalStatementMatchesAction,
					evalStatementMatchesPrincipal,
					evalStatementMatchesResource,
					evalStatementMatchesCondition,
				),
			)
			s.trc.Pop()
		}

		if !layerDecision.Allowed() {
			s.trc.Log("deny due to RCPs for node: %s of type %s", node.Name, node.Type)
			s.trc.Pop()
			return layerDecision
		}

		decision.Merge(layerDecision)
		s.trc.Pop()
	}

	return decision
}
