package sim

import (
	"strings"

	"github.com/nsiow/yams/pkg/policy"
)

var rcpSupportedServices = map[string]bool{
	"aoss":              true,
	"appconfig":         true,
	"appstream":         true,
	"autoscaling":       true,
	"codebuild":         true,
	"codecommit":        true,
	"cognito-identity":  true,
	"cognito-idp":       true,
	"cognito-sync":      true,
	"comprehend":        true,
	"comprehendmedical": true,
	"dax":               true,
	"dynamodb":          true,
	"ecr":               true,
	"health":            true,
	"kinesisvideo":      true,
	"kms":               true,
	"logs":              true,
	"s3":                true,
	"secretsmanager":    true,
	"signin":            true,
	"sqs":               true,
	"sts":               true,
	"support":           true,
	"textract":          true,
	"transcribe":        true,
	"translate":         true,
}

// supportsRCPs determines whether or not the provided auth context has support for RCPs based on:
// https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_rcps.html#rcp-supported-services
func supportsRCPs(s *subject) bool {
	return s.auth.Action != nil &&
		s.auth.Resource != nil &&
		rcpSupportedServices[s.auth.Action.Service] &&
		s.auth.Action.Targets(s.auth.Resource.Arn)
}

// evalRCP assesses the resource control policies of the Resource to determine whether or not it
// allows the provided AuthContext
func evalRCP(s *subject) Decision {
	trc := s.trc.Enabled()
	if trc {
		s.trc.Push("evaluating resource control policies")
		defer s.trc.Pop()
	}

	decision := Decision{}

	// Missing resource or empty RCP = allowed; otherwise we have to evaluate. Inspect every
	// node, not just the first, so an account-level RCP isn't skipped when the root has none.
	hasAnyRCP := false
	if s.auth.Resource != nil {
		for nodeIndex := range s.auth.Resource.Account.OrgNodes {
			node := &s.auth.Resource.Account.OrgNodes[nodeIndex]
			if len(node.RCPs) > 0 {
				hasAnyRCP = true
				break
			}
		}
	}
	if s.auth.Resource == nil ||
		len(s.auth.Resource.Account.OrgNodes) == 0 ||
		!hasAnyRCP {
		if trc {
			s.trc.Log("skipping RCPs: none found")
		}
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	if principalIsServiceLinkedRole(s) || resourceIsInManagementAccount(s) || rcpActionIsExempt(s) {
		if trc {
			s.trc.Log("skipping RCPs: exempt principal, resource, or action")
		}
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	// If service does not support RCPs, always allowed
	if !supportsRCPs(s) {
		if trc {
			s.trc.Log("action/resource does not support RCPs")
		}
		decision.Add(policy.EFFECT_ALLOW)
		return decision
	}

	// Iterate through layers of RCP, only continuing if we get an allow result through each layer
	for nodeIndex := range s.auth.Resource.Account.OrgNodes {
		node := &s.auth.Resource.Account.OrgNodes[nodeIndex]
		if trc {
			s.trc.Push("evaluating resource control policies for node: %s of type %s", node.Name, node.Type)
		}
		layerDecision := Decision{}

		for rcpIndex := range node.RCPs {
			rcp := &node.RCPs[rcpIndex]
			if trc {
				s.trc.Push("evaluating resource control policy: %s", rcp.Name)
			}

			localDecision := evalPolicy(s, &rcp.Policy,
				evalStatementMatchesAction,
				evalStatementMatchesPrincipal,
				evalStatementMatchesResource,
				evalStatementMatchesCondition,
			)
			if trc && localDecision.DeniedExplicit() {
				s.trc.Denied("explicit deny in resource control policy: %s", rcp.Name)
			}

			// Calculate access for this layer
			layerDecision.Merge(localDecision)

			if trc {
				s.trc.Pop()
			}
		}

		if !layerDecision.Allowed() {
			if trc {
				s.trc.Log("deny due to RCPs for node: %s of type %s", node.Name, node.Type)
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

func rcpActionIsExempt(s *subject) bool {
	return s.auth.Action != nil && strings.EqualFold(s.auth.Action.ShortName(), "kms:retiregrant")
}
