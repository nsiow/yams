package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalFunction is the blueprint of a function that allows us to evaluate a single statement
type evalFunction func(*subject, *policy.Statement) bool

// evalIsSameAccount determines whether or not the provided Principal + Resource exist within the
// same AWS account
func evalIsSameAccount(s *subject) bool {
	return s.auth.Resource == nil || s.auth.Principal.AccountId == s.auth.Resource.AccountId
}

// evalOverallAccess calculates both Principal + Resource access same performs both same-account
// and different-account evaluations
func evalOverallAccess(s *subject) *SimResult {

	// TODO(nsiow) revisit this ordering for accuracy vs speed tradeoffs
	// Calculate Resource access
	rAccess := evalResourceAccess(s)
	if rAccess.DeniedExplicit() {
		s.trc.Denied("[explicit deny] in resource policy")
		return &SimResult{Trace: &s.trc, IsAllowed: false}
	}
	if !rAccess.Allowed() && !evalIsSameAccount(s) {
		s.trc.Denied("[implicit deny] x-account, missing resource policy access")
	}

	// Calculate Principal access
	pAccess := evalPrincipalAccess(s)
	if pAccess.DeniedExplicit() {
		s.trc.Denied("[explicit deny] in identity policy")
		return &SimResult{Trace: &s.trc, IsAllowed: false}
	}
	if !pAccess.Allowed() && !evalIsSameAccount(s) {
		s.trc.Denied("[implicit deny] no identity-based policy allows access")
	}

	// Calculate SCP access, if present
	scpAccess := evalSCP(s)
	if scpAccess.DeniedExplicit() {
		s.trc.Denied("[explicit deny] in service control policies")
		return &SimResult{Trace: &s.trc, IsAllowed: false}
	}
	if !scpAccess.Allowed() {
		s.trc.Denied("[implicit deny] based on service control policies")
		return &SimResult{Trace: &s.trc, IsAllowed: false}
	}

	// Calculate RCP access, if present
	rcpAccess := evalRCP(s)
	if rcpAccess.DeniedExplicit() {
		s.trc.Denied("[explicit deny] in resource control policies")
		return &SimResult{Trace: &s.trc, IsAllowed: false}
	}
	if !rcpAccess.Allowed() {
		s.trc.Denied("[implicit deny] based on resource control policies")
		return &SimResult{Trace: &s.trc, IsAllowed: false}
	}

	// Calculate permissions boundary access, if present
	pbAccess := evalPermissionsBoundary(s)
	if pbAccess.DeniedExplicit() {
		s.trc.Denied("[explicit deny] in permissions boundary")
		return &SimResult{Trace: &s.trc, IsAllowed: false}
	}
	if !pbAccess.Allowed() {
		s.trc.Denied("[implicit deny] based on permissions boundary")
		return &SimResult{Trace: &s.trc, IsAllowed: false}
	}

	// If same account, access is granted if the Principal has access
	if evalIsSameAccount(s) {
		// Same-account-explicit-principal edge case
		if s.extra.ResourceAllowsExplicitPrincipal {
			s.trc.Allowed("[allow] access granted via same-account explicit-principal case")
			return &SimResult{Trace: &s.trc, IsAllowed: true}
		}

		if pAccess.Allowed() && !isStrictCall(s) {
			s.trc.Allowed("[allow] access granted via same-account identity policy")
			return &SimResult{Trace: &s.trc, IsAllowed: true}
		}

		if pAccess.Allowed() && rAccess.Allowed() {
			s.trc.Allowed("[allow] access granted via same-account identity policy (strict)")
			return &SimResult{Trace: &s.trc, IsAllowed: true}
		}

		if !pAccess.Allowed() {
			s.trc.Denied("[implicit deny] no identity-based policy allows this action")
			return &SimResult{Trace: &s.trc, IsAllowed: false}
		}
	}

	// Access is granted if the Principal has access and the Resource permits that access
	if pAccess.Allowed() && rAccess.Allowed() {
		s.trc.Allowed("[allow] access granted via x-account identity + resource policies")
		return &SimResult{Trace: &s.trc, IsAllowed: true}
	}
	if pAccess.Allowed() && !rAccess.Allowed() {
		s.trc.Denied("[implicit deny] x-account, missing resource policy access")
		return &SimResult{Trace: &s.trc, IsAllowed: false}
	}
	if !pAccess.Allowed() && rAccess.Allowed() {
		s.trc.Denied("[implicit deny] x-account, missing identity policy access")
		return &SimResult{Trace: &s.trc, IsAllowed: false}
	}

	// We fell through and no access was granted from either side
	s.trc.Denied("[implicit deny] x-account, missing both identity + resource access")
	return &SimResult{Trace: &s.trc, IsAllowed: false}
}
