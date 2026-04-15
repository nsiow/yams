package sim

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
		evalStatementMatchesPrincipal,
		evalStatementMatchesCondition,
	)

	// If the Principal and Resource are the same account, check whether the resource grants
	// access directly (vs delegating to the account) before returning
	s.extra.ResourceGrantsPrincipalAccess = evalResourceAccessGrantsPrincipal(s)

	return decision
}
