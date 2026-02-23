package sim

import (
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// evalPrincipalAccess calculates the Principal-side access to the specified Resource
func evalPrincipalAccess(s *subject) Decision {
	trc := s.trc.Enabled()
	if trc {
		s.trc.Push("evaluating principal policies")
		defer s.trc.Pop()
	}

	decision := Decision{}

	if len(s.auth.Principal.InlinePolicies) > 0 {
		if trc {
			s.trc.Push("evaluating inline principal policies")
		}
		decision.Merge(evalPrincipalHelperInline(s, "inline principal", s.auth.Principal.InlinePolicies))
		if trc {
			s.trc.Pop()
		}
	} else if trc {
		s.trc.Log("skipping inline policies: none found")
	}

	if len(s.auth.Principal.AttachedPolicies) > 0 {
		if trc {
			s.trc.Push("evaluating attached principal policies")
		}
		decision.Merge(evalPrincipalHelperAttached(s, "attached principal", s.auth.Principal.AttachedPolicies))
		if trc {
			s.trc.Pop()
		}
	} else if trc {
		s.trc.Log("skipping attached policies: none found")
	}

	if len(s.auth.Principal.Groups) > 0 {
		if trc {
			s.trc.Push("evaluating group-based principal policies")
		}
		decision.Merge(evalPrincipalGroupPolicies(s, s.auth.Principal.Groups))
		if trc {
			s.trc.Pop()
		}
	} else if trc {
		s.trc.Log("skipping group policies: none found")
	}

	return decision
}

// evalPrincipalGroupPolicies calculates the Principal-side access based on group policies
func evalPrincipalGroupPolicies(s *subject, groups []entities.FrozenGroup) Decision {
	trc := s.trc.Enabled()
	if trc {
		s.trc.Push("evaluating group principal policies")
		defer s.trc.Pop()
	}

	decision := Decision{}
	for _, group := range groups {
		if trc {
			s.trc.Push("evaluating inline policies for group: %s", group.Arn)
		}
		decision.Merge(evalPrincipalHelperInline(s, "inline group", group.InlinePolicies))
		if trc {
			s.trc.Pop()
		}

		if trc {
			s.trc.Push("evaluating attached policies for group: %s", group.Arn)
		}
		decision.Merge(evalPrincipalHelperAttached(s, "attached group", group.AttachedPolicies))
		if trc {
			s.trc.Pop()
		}
	}

	return decision
}

// evalPrincipalHelperInline is a helper function for easier evaluation of inline policies
func evalPrincipalHelperInline(s *subject, pType string, inline []policy.Policy) Decision {
	trc := s.trc.Enabled()
	decision := Decision{}
	for i, pol := range inline {
		if trc {
			// Prefer Name (inline policy name) over Id (policy document id)
			policyName := pol.Name
			if policyName == "" {
				policyName = Id(pol.Id, i)
			}
			s.trc.Push("evaluating %s policy: %s", pType, policyName)
		}

		localDecision := evalPolicy(s, pol,
			evalStatementMatchesAction,
			evalStatementMatchesResource,
			evalStatementMatchesCondition)

		if trc {
			if localDecision.Allowed() {
				policyName := pol.Name
				if policyName == "" {
					policyName = Id(pol.Id, i)
				}
				s.trc.Allowed("allow in %s policy: %s", pType, policyName)
			}
			if localDecision.DeniedExplicit() {
				policyName := pol.Name
				if policyName == "" {
					policyName = Id(pol.Id, i)
				}
				s.trc.Denied("explicit deny in %s policy: %s", pType, policyName)
			}
			s.trc.Pop()
		}

		decision.Merge(localDecision)
	}

	return decision
}

// evalPrincipalHelperAttached is a helper function for easier evaluation of attached policies
func evalPrincipalHelperAttached(s *subject, pType string, att []entities.ManagedPolicy) Decision {
	trc := s.trc.Enabled()
	decision := Decision{}
	for _, policy := range att {
		if trc {
			s.trc.Push("evaluating %s policy: %s", pType, policy.Arn)
		}

		localDecision := evalPolicy(s, policy.Policy,
			evalStatementMatchesAction,
			evalStatementMatchesResource,
			evalStatementMatchesCondition)

		if trc {
			if localDecision.Allowed() {
				s.trc.Allowed("allow in %s policy: %s", pType, policy.Arn)
			}
			if localDecision.DeniedExplicit() {
				s.trc.Denied("explicit deny in %s policy: %s", pType, policy.Arn)
			}
			s.trc.Pop()
		}

		decision.Merge(localDecision)
	}

	return decision
}
