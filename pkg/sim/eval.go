package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalFunction is the blueprint of a function that allows us to evaluate a single statement
type evalFunction func(*subject, *policy.Statement) bool

// evalIsSameAccount determines whether or not the provided Principal + Resource exist within the
// same AWS account
func evalIsSameAccount(s *subject) bool {
	return s.auth.Principal != nil &&
		s.auth.Resource != nil &&
		s.auth.Principal.AccountId == s.auth.Resource.AccountId
}

// evalOverallAccess calculates both Principal + Resource access same performs both same-account
// and different-account evaluations
func evalOverallAccess(s *subject) (*Result, error) {

	// Calculate SCP access, if present
	scpAccess := evalSCP(s)
	if scpAccess.Contains(policy.EFFECT_DENY) {
		s.trc.Decision("[explicit deny] found in service control policies")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}
	if !scpAccess.Allowed() {
		s.trc.Decision("[implicit deny] based on service control policies")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}

	// Calculate permissions boundary access, if present
	pbAccess := evalPermissionsBoundary(s)
	if pbAccess.Contains(policy.EFFECT_DENY) {
		s.trc.Decision("[explicit deny] found in permissions boundary")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}
	if !pbAccess.Allowed() {
		s.trc.Decision("[implicit deny] based on permissions boundary")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}

	// Calculate Principal access
	pAccess := evalPrincipalAccess(s)
	// ... check for explicit Deny results
	if pAccess.Contains(policy.EFFECT_DENY) {
		s.trc.Decision("[explicit deny] found in identity policy")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}

	// Calculate Resource access
	rAccess := evalResourceAccess(s)
	// ... check for explicit Deny results
	if rAccess.Contains(policy.EFFECT_DENY) {
		s.trc.Decision("[explicit deny] found in resource policy")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}

	// If same account, access is granted if the Principal has access
	if evalIsSameAccount(s) {
		if pAccess.Contains(policy.EFFECT_ALLOW) && !isStrictCall(s) {
			s.trc.Decision("[allow] access granted via same-account identity policy")
			return &Result{Trace: s.trc, IsAllowed: true}, nil
		}

		// Same-account-explicit-principal edge case
		if s.extra.ResourceAllowsExplicitPrincipal {
			s.trc.Decision("[allow] access granted via same-account explicit-principal case")
			return &Result{Trace: s.trc, IsAllowed: true}, nil
		}

		s.trc.Decision("[implicit deny] no identity-based policy allows this action")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}

	// Access is granted if the Principal has access and the Resource permits that access
	if pAccess.Contains(policy.EFFECT_ALLOW) && rAccess.Contains(policy.EFFECT_ALLOW) {
		s.trc.Decision("[allow] access granted via x-account identity + resource policies")
		return &Result{Trace: s.trc, IsAllowed: true}, nil
	}
	if pAccess.Contains(policy.EFFECT_ALLOW) && !rAccess.Contains(policy.EFFECT_ALLOW) {
		s.trc.Decision("[implicit deny] x-account, missing resource policy access")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}
	if !pAccess.Contains(policy.EFFECT_ALLOW) && rAccess.Contains(policy.EFFECT_ALLOW) {
		s.trc.Decision("[implicit deny] x-account, missing identity policy access")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}

	// We fell through and no access was granted from either side
	s.trc.Decision("[implicit deny] x-account, missing both identity + resource access")
	return &Result{Trace: s.trc, IsAllowed: false}, nil
}
