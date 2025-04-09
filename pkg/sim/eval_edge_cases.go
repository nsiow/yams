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
