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
func evalOverallAccess(s *subject) (*SimResult, error) {

	// Calculate SCP access, if present
	scpAccess := evalSCP(s)
	if scpAccess.DeniedExplicit() {
		s.trc.Denied("[explicit deny] found in service control policies")
		return &SimResult{Trace: s.trc, IsAllowed: false}, nil
	}
	if !scpAccess.Allowed() {
		s.trc.Denied("[implicit deny] based on service control policies")
		return &SimResult{Trace: s.trc, IsAllowed: false}, nil
	}
	// Calculate RCP access, if present
	rcpAccess := evalRCP(s)
	if rcpAccess.DeniedExplicit() {
		s.trc.Denied("[explicit deny] found in resource control policies")
		return &SimResult{Trace: s.trc, IsAllowed: false}, nil
	}
	if !rcpAccess.Allowed() {
		s.trc.Denied("[implicit deny] based on resource control policies")
		return &SimResult{Trace: s.trc, IsAllowed: false}, nil
	}

	// Calculate permissions boundary access, if present
	pbAccess := evalPermissionsBoundary(s)
	if pbAccess.DeniedExplicit() {
		s.trc.Denied("[explicit deny] found in permissions boundary")
		return &SimResult{Trace: s.trc, IsAllowed: false}, nil
	}
	if !pbAccess.Allowed() {
		s.trc.Denied("[implicit deny] based on permissions boundary")
		return &SimResult{Trace: s.trc, IsAllowed: false}, nil
	}

	// Calculate Principal access
	pAccess := evalPrincipalAccess(s)
	if pAccess.DeniedExplicit() {
		s.trc.Denied("[explicit deny] found in identity policy")
		return &SimResult{Trace: s.trc, IsAllowed: false}, nil
	}

	// Calculate Resource access
	rAccess := evalResourceAccess(s)
	if rAccess.DeniedExplicit() {
		s.trc.Denied("[explicit deny] found in resource policy")
		return &SimResult{Trace: s.trc, IsAllowed: false}, nil
	}

	// If same account, access is granted if the Principal has access
	if evalIsSameAccount(s) {
		if pAccess.Allow && !isStrictCall(s) {
			s.trc.Allowed("[allow] access granted via same-account identity policy")
			return &SimResult{Trace: s.trc, IsAllowed: true}, nil
		}

		// Same-account-explicit-principal edge case
		if s.extra.ResourceAllowsExplicitPrincipal {
			s.trc.Allowed("[allow] access granted via same-account explicit-principal case")
			return &SimResult{Trace: s.trc, IsAllowed: true}, nil
		}

		s.trc.Denied("[implicit deny] no identity-based policy allows this action")
		return &SimResult{Trace: s.trc, IsAllowed: false}, nil
	}

	// Access is granted if the Principal has access and the Resource permits that access
	if pAccess.Allow && rAccess.Allow {
		s.trc.Allowed("[allow] access granted via x-account identity + resource policies")
		return &SimResult{Trace: s.trc, IsAllowed: true}, nil
	}
	if pAccess.Allow && !rAccess.Allow {
		s.trc.Denied("[implicit deny] x-account, missing resource policy access")
		return &SimResult{Trace: s.trc, IsAllowed: false}, nil
	}
	if !pAccess.Allow && rAccess.Allow {
		s.trc.Denied("[implicit deny] x-account, missing identity policy access")
		return &SimResult{Trace: s.trc, IsAllowed: false}, nil
	}

	// We fell through and no access was granted from either side
	s.trc.Denied("[implicit deny] x-account, missing both identity + resource access")
	return &SimResult{Trace: s.trc, IsAllowed: false}, nil
}
