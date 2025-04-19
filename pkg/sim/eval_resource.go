package sim

// evalResourceAccess calculates the Resource-side access with regard to the specified Principal
func evalResourceAccess(s *subject) Decision {

	s.trc.Push("evaluating resource policies")
	defer s.trc.Pop()

	// Iterate over resource policy statements to evaluate access
	decision := evalPolicy(s,
		s.ac.Resource.Policy,
		evalStatementMatchesAction,
		evalStatementMatchesPrincipal,
		evalStatementMatchesCondition,
	)

	// If the Principal and Resource are the same account, check for the explicit-principal edge
	// case before returning
	if evalIsSameAccount(s) && !s.ac.Resource.Policy.Empty() {
		edgeCaseDecision := evalPolicy(s, s.ac.Resource.Policy, evalStatementMatchesPrincipalExact)
		if edgeCaseDecision.Allowed() {
			s.extra.ResourceAllowsExplicitPrincipal = true
		}
	}

	return decision
}
