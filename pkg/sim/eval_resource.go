package sim

// evalResourceAccess calculates the Resource-side access with regard to the specified Principal
func evalResourceAccess(s *subject) Decision {

	s.trc.Push("evaluating resource policies")
	defer s.trc.Pop()

	// Iterate over resource policy statements to evaluate access
	// FIXME(nsiow) this also needs evalStatementMatchesResource
	decision := evalPolicy(s, s.ac.Resource.FrozenPolicy,
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesPrincipal,
		evalStatementMatchesCondition,
	)

	// If the Principal and Resource are the same account, check for the explicit-principal edge
	// case before returning
	s.extra.ResourceAllowsExplicitPrincipal = evalResourceAccessAllowsExplicitPrincipal(s)

	return decision
}
