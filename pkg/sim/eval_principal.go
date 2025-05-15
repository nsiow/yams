package sim

import (
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// evalPrincipalAccess calculates the Principal-side access to the specified Resource
func evalPrincipalAccess(s *subject) Decision {
	s.trc.Push("evaluating principal policies")
	defer s.trc.Pop()

	decision := Decision{}

	if len(s.auth.Principal.InlinePolicies) > 0 {
		s.trc.Push("evaluating inline principal policies")
		decision.Merge(evalPrincipalHelperInline(s, "inline principal", s.auth.Principal.InlinePolicies))
		s.trc.Pop()
	} else {
		s.trc.Log("skipping inline policies: none found")
	}

	if len(s.auth.Principal.AttachedPolicies) > 0 {
		s.trc.Push("evaluating attached principal policies")
		decision.Merge(evalPrincipalHelperAttached(s, "attached principal", s.auth.Principal.AttachedPolicies))
		s.trc.Pop()
	} else {
		s.trc.Log("skipping attached policies: none found")
	}

	if len(s.auth.Principal.Groups) > 0 {
		s.trc.Push("evaluating group-based principal policies")
		decision.Merge(evalPrincipalGroupPolicies(s, s.auth.Principal.Groups))
		s.trc.Pop()
	} else {
		s.trc.Log("skipping group policies: none found")
	}

	return decision
}

// evalPrincipalGroupPolicies calculates the Principal-side access based on group policies
func evalPrincipalGroupPolicies(s *subject, groups []entities.FrozenGroup) Decision {
	s.trc.Push("evaluating group principal policies")
	defer s.trc.Pop()

	decision := Decision{}
	for _, group := range groups {
		s.trc.Push("evaluating inline policies for group: %s", group.Arn)
		decision.Merge(evalPrincipalHelperInline(s, "inline group", group.InlinePolicies))
		s.trc.Pop()

		s.trc.Push("evaluating attached policies for group: %s", group.Arn)
		decision.Merge(evalPrincipalHelperAttached(s, "attached group", group.AttachedPolicies))
		s.trc.Pop()
	}

	return decision
}

// evalPrincipalHelperInline is a helper function for easier evaluation of inline policies
func evalPrincipalHelperInline(s *subject, pType string, inline []policy.Policy) Decision {
	decision := Decision{}
	for i, policy := range inline {
		s.trc.Push("evaluating %s policy: %s", pType, Id(policy.Id, i))
		localDecision := evalPolicy(s, policy,
			evalStatementMatchesAction,
			evalStatementMatchesResource,
			evalStatementMatchesCondition)

		if localDecision.Allowed() {
			s.trc.Allowed("allow in %s policy: %s", pType, Id(policy.Id, i))
		}
		if localDecision.DeniedExplicit() {
			s.trc.Denied("explicit deny in %s policy: %s", pType, Id(policy.Id, i))
		}

		decision.Merge(localDecision)
		s.trc.Pop()
	}

	return decision
}

// evalPrincipalInlinePolicies is a helper function for easier evaluation of inline policies
func evalPrincipalHelperAttached(s *subject, pType string, att []entities.ManagedPolicy) Decision {
	decision := Decision{}
	for _, policy := range att {
		s.trc.Push("evaluating %s policy: %s", pType, policy.Arn)
		localDecision := evalPolicy(s, policy.Policy,
			evalStatementMatchesAction,
			evalStatementMatchesResource,
			evalStatementMatchesCondition)

		if localDecision.Allowed() {
			s.trc.Allowed("allow in %s policy: %s", pType, policy.Arn)
		}
		if localDecision.DeniedExplicit() {
			s.trc.Denied("explicit deny in %s policy: %s", pType, policy.Arn)
		}

		decision.Merge(localDecision)
		s.trc.Pop()
	}

	return decision
}
