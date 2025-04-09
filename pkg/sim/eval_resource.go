package sim

// evalResourceAccess calculates the Resource-side access with regard to the specified Principal
func evalResourceAccess(s *subject) (Decision, error) {

	s.trc.Push("evaluating resource policies")
	defer s.trc.Pop()

	// Specify the statement evaluation funcs we will consider for Principal access
	funcs := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesPrincipal,
		evalStatementMatchesCondition,
	}

	// Iterate over resource policy statements to evaluate access
	dec, err := evalPolicy(s, s.ac.Resource.Policy, funcs)
	if err != nil {
		return dec, err
	}

	// If the Principal and Resource are the same account, check for the explicit-principal edge
	// case before returning
	if evalIsSameAccount(s) && !s.ac.Resource.Policy.Empty() {
		funcs = []evalFunction{evalStatementMatchesPrincipalExact}
		edgeCaseDecision, err := evalPolicy(s, s.ac.Resource.Policy, funcs)
		if err != nil {
			return dec, err
		}
		if edgeCaseDecision.Allowed() {
			s.extra.ResourceAllowsExplicitPrincipal = true
		}
	}

	return dec, nil
}
