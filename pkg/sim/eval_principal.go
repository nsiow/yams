package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// evalPrincipalAccess calculates the Principal-side access to the specified Resource
func evalPrincipalAccess(s *subject) Decision {
	s.trc.Push("evaluating principal policies")
	defer s.trc.Pop()

	decision := Decision{}

	s.trc.Push("evaluating inline principal policies")
	decision.Merge(evalPrincipalHelperInline(s, s.ac.Principal.InlinePolicies))
	s.trc.Pop()

	s.trc.Push("evaluating attached principal policies")
	decision.Merge(evalPrincipalHelperAttached(s, s.ac.Principal.FrozenAttachedPolicies))
	s.trc.Pop()

	s.trc.Push("evaluating group-based principal policies")
	decision.Merge(evalPrincipalGroupPolicies(s, s.ac.Principal.FrozenGroups))
	s.trc.Pop()

	return decision
}

// evalPrincipalGroupPolicies calculates the Principal-side access based on group policies
func evalPrincipalGroupPolicies(s *subject, groups []entities.FrozenGroup) Decision {
	s.trc.Push("evaluating group principal policies")
	defer s.trc.Pop()

	decision := Decision{}
	for _, group := range groups {
		s.trc.Push("evaluating inline group principal policies for group: %s", group.Arn)
		decision.Merge(evalPrincipalHelperInline(s, group.InlinePolicies))
		s.trc.Pop()

		s.trc.Push("evaluating attached group principal policies for group: %s", group.Arn)
		decision.Merge(evalPrincipalHelperAttached(s, group.FrozenAttachedPolicies))
		s.trc.Pop()
	}

	return decision
}

// evalPrincipalHelperInline is a helper function for easier evaluation of inline policies
func evalPrincipalHelperInline(s *subject, inline []policy.Policy) (decision Decision) {
	for i, policy := range inline {
		var pid string
		if len(policy.Id) > 0 {
			pid = policy.Id
		} else {
			pid = fmt.Sprintf("inline(%d)", i)
		}

		s.trc.Push("evaluating inline policy: %s", pid)
		defer s.trc.Pop()

		return evalPolicy(s, policy,
			evalStatementMatchesAction,
			evalStatementMatchesResource,
			evalStatementMatchesCondition)
	}

	return decision
}

// evalPrincipalInlinePolicies is a helper function for easier evaluation of inline policies
func evalPrincipalHelperAttached(s *subject, attached []entities.ManagedPolicy) Decision {
	decision := Decision{}

	for _, policy := range attached {
		s.trc.Push("evaluating inline policy: %s", policy.Arn)
		decision.Merge(
			evalPolicy(s, policy.Policy,
				evalStatementMatchesAction,
				evalStatementMatchesResource,
				evalStatementMatchesCondition))
		s.trc.Pop()
	}

	return decision
}
