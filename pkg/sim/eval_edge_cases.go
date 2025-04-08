package sim

import (
	"strings"
)

// isStrictCall returns whether the specified API is one that requires both Principal + Resource
// policy in order to be allowed; even if same account
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

// evalSameAccountExplicitPrincipalCase handles the special case where the Resource policy
// granting explicit access to the Principal circumvents the need for Principal-policy access
func evalSameAccountExplicitPrincipalCase(s *subject) (Decision, error) {

	// Has to have a resource policy, otherwise this will never have an effect
	if s.ac.Resource == nil || s.ac.Resource.Policy.Empty() {
		return Decision{}, nil
	}

	s.trc.Push("evaluating same-account explicit-resource-policy edge case")
	defer s.trc.Pop()

	// Specify the statement evaluation funcs we will consider for Principal access
	funcs := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesPrincipalExact,
		evalStatementMatchesCondition,
	}

	// Iterate over resource policy statements to evaluate access
	return evalPolicy(s, s.ac.Resource.Policy, funcs)
}
