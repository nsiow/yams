package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
)

// evalFunction is the blueprint of a function that allows us to evaluate a single statement
type evalFunction func(*subject, *policy.Statement) (bool, error)

// evalIsSameAccount determines whether or not the provided Principal + Resource exist within the
// same AWS account
func evalIsSameAccount(s *subject) bool {
	return s.ac.Principal != nil &&
		s.ac.Resource != nil &&
		s.ac.Principal.AccountId == s.ac.Resource.AccountId
}

// evalOverallAccess calculates both Principal + Resource access same performs both same-account
// and different-account evaluations
func evalOverallAccess(s *subject) (*Result, error) {

	// Calculate SCP access, if present
	scpAccess, err := evalSCP(s)
	if err != nil {
		return nil, fmt.Errorf("error evaluating SCP: %w", err)
	}
	if scpAccess.Contains(policy.EFFECT_DENY) {
		s.trc.Decision("[explicit deny] found in service control policies")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}
	if !scpAccess.Allowed() {
		s.trc.Decision("[implicit deny] based on service control policies")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}

	// Calculate permissions boundary access, if present
	pbAccess, err := evalPermissionsBoundary(s)
	if err != nil {
		return nil, fmt.Errorf("error evaluating permission boundary: %w", err)
	}
	if pbAccess.Contains(policy.EFFECT_DENY) {
		s.trc.Decision("[explicit deny] found in permissions boundary")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}
	if !pbAccess.Allowed() {
		s.trc.Decision("[implicit deny] based on permissions boundary")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}

	// Calculate Principal access
	pAccess, err := evalPrincipalAccess(s)
	if err != nil {
		return nil, fmt.Errorf("error evaluating principal access: %w", err)
	}
	// ... check for explicit Deny results
	if pAccess.Contains(policy.EFFECT_DENY) {
		s.trc.Decision("[explicit deny] found in identity policy")
		return &Result{Trace: s.trc, IsAllowed: false}, nil
	}

	// Calculate Resource access
	rAccess, extra, err := evalResourceAccess(s)
	if err != nil {
		return nil, fmt.Errorf("error evaluating resource access: %w", err)
	}
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
		if extra.ResourceAccessAllowsExplicitPrincipal {
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
