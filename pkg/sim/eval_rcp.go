package sim

import (
	"strings"

	"github.com/nsiow/yams/pkg/entities"
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
	if s.auth.Resource == nil || len(s.auth.Resource.Account.RCPs) == 0 {
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
	rcps := s.auth.Resource.Account.RCPs
	for i, layer := range rcps {

		s.trc.Push("evaluating SCP layer %d of %d", i, len(rcps)-1)

		// Calculate access for this layer
		decision = evalRCPLayer(s, layer)

		// If not allowed at this layer, propagate result up; should be denied
		if !decision.Allowed() {
			s.trc.Pop()
			return decision
		}

		s.trc.Pop()
	}

	return decision
}

// evalRCPLayer evaluates a single "layer" of resource control policies
//
// This is separated since each logical layer must result in an ALLOW decision in order to
// continue
func evalRCPLayer(s *subject, layer []entities.ManagedPolicy) (decision Decision) {
	for _, pol := range layer {
		s.trc.Push("evaluating RCP: %s", pol.Arn)
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
