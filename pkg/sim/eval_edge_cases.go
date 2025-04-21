package sim

import (
	"strings"
)

// isStrictCall returns whether the specified API is one that requires both Principal + Resource
// policy in order to be allowed; even if same account
// TODO(nsiow) reimplement x-account logic to just reuse this code path rather than having two
// different trees
// TODO(nsiow) this should just be a method on Action or something
func isStrictCall(s *subject) bool {
	// strict calls always require involve both a Principal + Resource
	if s.ac == nil || s.ac.Principal == nil || s.ac.Resource == nil {
		return false
	}

	// sts assume-role case
	if strings.EqualFold("sts:assumerole", s.ac.Action.ShortName()) {
		return true
	}

	// kms case
	if strings.EqualFold("kms", s.ac.Action.Service) &&
		strings.EqualFold("AWS::KMS::Key", s.ac.Resource.Type) {
		return true
	}

	// most calls are not strict by default
	return false
}

// evalResourceAccessAllowsExplicitPrincipal tests for the edge case where a same-account resource
// allows a Principal by ARN specifically, which has an effect on evaluation logic
func evalResourceAccessAllowsExplicitPrincipal(s *subject) bool {
	if evalIsSameAccount(s) && !s.ac.Resource.FrozenPolicy.Empty() {
		subDecision := evalPolicy(s, s.ac.Resource.FrozenPolicy,
			evalStatementMatchesAction,
			evalStatementMatchesPrincipalExact,
			evalStatementMatchesCondition)
		if subDecision.Allowed() {
			return true
		}
	}

	return false
}
