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
	if s.auth.Principal == nil || s.auth.Resource == nil {
		return false
	}

	// sts assume-role case
	if strings.EqualFold("sts:assumerole", s.auth.Action.ShortName()) {
		return true
	}

	// kms case
	if strings.EqualFold("kms", s.auth.Action.Service) &&
		strings.EqualFold("AWS::KMS::Key", s.auth.Resource.Type) {
		return true
	}

	// most calls are not strict by default
	return false
}

// evalResourceAccessGrantsPrincipal tests for the edge case where a same-account resource policy
// grants a principal access directly (not via delegation). When this is true, the principal does
// not need identity policies to access the resource.
//
// Delegation (account-ID or account-root in the Principal block) is excluded because it just
// defers access decisions to the account's identity policies.
func evalResourceAccessGrantsPrincipal(s *subject) bool {
	s.trc.Push("evaluating whether the resource grants the principal access directly")
	defer s.trc.Pop()

	if s.auth.Resource == nil {
		return false
	}

	if evalIsSameAccount(s) && !s.auth.Resource.Policy.Empty() {
		subDecision := evalPolicy(s, s.auth.Resource.Policy,
			evalStatementMatchesAction,
			evalStatementMatchesPrincipal,
			evalStatementIsNotDelegated,
			evalStatementMatchesCondition)
		if subDecision.Allowed() {
			return true
		}
	}

	return false
}
