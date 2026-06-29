package sim

import (
	"strings"

	"github.com/nsiow/yams/pkg/loaders/awsconfig"
	"github.com/nsiow/yams/pkg/policy"
)

// evalResourceAccess calculates the Resource-side access with regard to the specified Principal
func evalResourceAccess(s *subject) Decision {
	trc := s.trc.Enabled()
	if trc {
		s.trc.Push("evaluating resource policies")
		defer s.trc.Pop()
	}

	if s.auth.Resource == nil || s.auth.Resource.Policy.Empty() {
		if trc {
			s.trc.Log("skipping resource policy: none found")
		}
		return Decision{}
	}

	// Iterate over resource policy statements to evaluate access
	decision := evalPolicy(s, s.auth.Resource.Policy,
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalResourceStatementMatchesPrincipal,
		evalStatementMatchesCondition,
	)

	// If the Principal and Resource are the same account, check whether the resource grants
	// access directly (vs delegating to the account) before returning
	s.extra.ResourceGrantsPrincipalAccess = evalResourceAccessGrantsPrincipal(s)

	return decision
}

func evalResourceStatementMatchesPrincipal(s *subject, stmt *policy.Statement) bool {
	if evalResourceDenyNotPrincipalMatchesBoundary(s, stmt) {
		return true
	}
	return evalStatementMatchesPrincipal(s, stmt)
}

func evalResourceDenyNotPrincipalMatchesBoundary(s *subject, stmt *policy.Statement) bool {
	return stmt.Effect == policy.EFFECT_DENY &&
		(stmt.NotPrincipal.All || !stmt.NotPrincipal.Empty()) &&
		evalPrincipalHasPermissionsBoundary(s) &&
		evalPrincipalIsIAMUserOrRole(s)
}

func evalResourceGrantRequiresBoundary(s *subject) bool {
	return s.auth.Principal != nil &&
		strings.EqualFold(s.auth.Principal.Type, awsconfig.CONST_TYPE_AWS_IAM_ROLE)
}

func evalPrincipalHasPermissionsBoundary(s *subject) bool {
	return s.auth.Principal != nil && s.auth.Principal.PermissionBoundary.Arn != ""
}

func evalPrincipalIsIAMUserOrRole(s *subject) bool {
	if s.auth.Principal == nil {
		return false
	}
	return strings.EqualFold(s.auth.Principal.Type, awsconfig.CONST_TYPE_AWS_IAM_USER) ||
		strings.EqualFold(s.auth.Principal.Type, awsconfig.CONST_TYPE_AWS_IAM_ROLE)
}
