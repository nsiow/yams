package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
)

// evalPrincipalAccess calculates the Principal-side access to the specified Resource
func evalPrincipalAccess(s *subject) Decision {

	s.trc.Push("evaluating principal policies")
	defer s.trc.Pop()

	// Specify the types of policies we will consider for Principal access
	effectivePolicies := map[string][]policy.Policy{
		"inline":  s.ac.Principal.InlinePolicies,
		"managed": s.ac.Principal.AttachedPolicies,
		"group":   s.ac.Principal.GroupPolicies,
	}

	// Specify the statement evaluation funcs we will consider for Principal access
	funcs := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesCondition,
	}

	// Iterate over policy types / policies / statements to evaluate access
	decision := Decision{}
	for policytype, policies := range effectivePolicies {
		s.trc.Push(fmt.Sprintf("policytype=%s", policytype))
		effect := evalPolicies(s, policies, funcs)
		decision.Merge(effect)
		s.trc.Pop()
	}

	return decision
}
