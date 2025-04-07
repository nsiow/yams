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
	return evalPolicy(s, s.ac.Resource.Policy, funcs)
}
