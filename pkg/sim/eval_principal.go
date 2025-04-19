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

	decision.Merge(
		evalPrincipalInlinePolicies(s, s.ac.Principal.InlinePolicies),
		evalPrincipalAttachedPolicies(s, s.ac.Principal.FrozenAttachedPolicies),
		evalPrincipalGroupPolicies(s, s.ac.Principal.FrozenGroups),
	)

	return decision
}

// evalPrincipalInlinePolicies calculates the Principal-side access based on inline policies
func evalPrincipalInlinePolicies(s *subject, policies []policy.Policy) Decision {
	s.trc.Push("evaluating principal inline policies")
	defer s.trc.Pop()

	decision := Decision{}
	for i, policy := range policies {
		var policyId string
		if len(policy.Id) > 0 {
			policyId = policyId
		} else {
			policyId = fmt.Sprintf("inline(%d)", i)
		}

		s.trc.Push(fmt.Sprintf("evaluating principal inline policy: %s", policyId))
		decision.Merge(
			evalPolicy(
				s,
				policy,
				evalStatementMatchesAction,
				evalStatementMatchesResource,
				evalStatementMatchesCondition,
			),
		)
		s.trc.Pop()
	}

	return decision
}

// evalPrincipalAttachedPolicies calculates the Principal-side access based on attached policies
func evalPrincipalAttachedPolicies(s *subject, policies []entities.ManagedPolicy) Decision {
	s.trc.Push("evaluating principal attached policies")
	defer s.trc.Pop()

	decision := Decision{}
	for _, policy := range policies {
		s.trc.Push(fmt.Sprintf("evaluating principal attached policy: %s", policy.Arn))
		decision.Merge(
			evalPolicy(
				s,
				policy.Policy,
				evalStatementMatchesAction,
				evalStatementMatchesResource,
				evalStatementMatchesCondition,
			),
		)
		s.trc.Pop()
	}

	return decision
}

// evalPrincipalGroupPolicies calculates the Principal-side access based on group policies
func evalPrincipalGroupPolicies(s *subject, groups []entities.FrozenGroup) Decision {
	s.trc.Push("evaluating group principal policies")
	defer s.trc.Pop()

	decision := Decision{}
	for _, group := range groups {
		// FIXME(nsiow) move trc.Push(fmt.Sprintf(...)) to use new args format
		s.trc.Push("evaluating group principal policies for group: %s", group.Arn)
		for _, policy := range group.FrozenPolicies {
			s.trc.Push("evaluating group principal policy: %s", policy.Arn)
			decision.Merge(
				evalPolicy(
					s,
					policy.Policy,
					evalStatementMatchesAction,
					evalStatementMatchesResource,
					evalStatementMatchesCondition,
				),
			)
			s.trc.Pop()
		}
		s.trc.Pop()
	}

	return decision
}
