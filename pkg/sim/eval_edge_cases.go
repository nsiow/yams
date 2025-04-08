package sim

import "strings"

// isStrictCall returns whether the specified API is one that requires both Principal + Resource
// policy in order to be allowed; even if same account
func isStrictCall(s *subject) bool {
	// strict calls always require involve both a Principal + Resource
	if s.ac.Principal == nil || s.ac.Resource == nil {
		return false
	}

	// sts assume-role case
	if strings.EqualFold("sts:assumerole", s.ac.Action.ShortName()) {
		return true
	}

	// kms case
	if strings.EqualFold("kms", s.ac.Action.ShortName()) &&
		strings.EqualFold("AWS::KMS::Key", s.ac.Resource.Type) {
		return true
	}

	// most calls are not strict by default
	return false
}

// evalSameAccountExplicitPrincipalCase handles the special case where the Resource policy
// granting explicit access to the Principal circumvents the need for Principal-policy access
// TODO(nsiow) figure out
func evalSameAccountExplicitPrincipalCase(s *subject) bool {
	// this edge case always requires that principal + resource exist in the same account, and to do
	// so they must actually both exist
	if s.ac.Principal == nil || s.ac.Resource == nil {
		return false
	}

	s.trc.Push("evaluating same-account explicit-resource-policy edge case")
	defer s.trc.Pop()

	for _, stmt := range s.ac.Resource.Policy.Statement {
		if stmt.Principal.AWS.Contains(s.ac.Principal.Arn) {
			return true
		}
	}

	return false
}
