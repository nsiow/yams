package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/trace"
)

// evalFunction is the blueprint of a function that allows us to evaluate a single statement
type evalFunction func(*subject, *policy.Statement) (bool, error)

// evalIsSameAccount determines whether or not the provided Principal + Resource exist within the
// same AWS account
func evalIsSameAccount(p *entities.Principal, r *entities.Resource) bool {
	return p.AccountId == r.AccountId
}

// evalSameAccountExplicitPrincipalCase handles the special case where the Resource policy
// granting explicit access to the Principal circumvents the need for Principal-policy access
// func evalSameAccountExplicitPrincipalCase(_ *entities.Principal, _ *entities.Resource) bool {
// 	// TODO(nsiow) implement correct behavior for same-account access via explicit ARN
// 	return false
// }

// evalOverallAccess calculates both Principal + Resource access same performs both same-account
// and different-account evaluations
func evalOverallAccess(s *subject) (*Result, error) {

	// FIXME(nsiow) figure out where trace should be set -- here or when `subject` is created
	trc := trace.New()

	// TODO(nsiow) good time to validate that the ac Action even applies to the ac Resource based on
	//             SAR values

	// TODO(nsiow) this may be ridiculously too large to include in trace
	trc.Attr("authContext", s.ac)

	// Calculate SCP access, if present
	scpAccess, err := evalSCP(s)
	if err != nil {
		return nil, fmt.Errorf("error evaluating SCP: %w", err)
	}
	if scpAccess.Contains(policy.EFFECT_DENY) {
		trc.Decision("[explicit deny] found in service control policies")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}
	if !scpAccess.Allowed() {
		trc.Decision("[implicit deny] based on service control policies")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}

	// Calculate permissions boundary access, if present
	pbAccess, err := evalPermissionsBoundary(s)
	if err != nil {
		return nil, fmt.Errorf("error evaluating permission boundary: %w", err)
	}
	if pbAccess.Contains(policy.EFFECT_DENY) {
		trc.Decision("[explicit deny] found in permissions boundary")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}
	if !pbAccess.Allowed() {
		trc.Decision("[implicit deny] based on permissions boundary")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}

	// Calculate Principal access
	pAccess, err := evalPrincipalAccess(s)
	if err != nil {
		return nil, fmt.Errorf("error evaluating principal access: %w", err)
	}
	// ... check for explicit Deny results
	if pAccess.Contains(policy.EFFECT_DENY) {
		trc.Decision("[explicit deny] found in identity policy")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}

	// Calculate Resource access
	rAccess, err := evalResourceAccess(s)
	if err != nil {
		return nil, fmt.Errorf("error evaluating resource access: %w", err)
	}
	// ... check for explicit Deny results
	if rAccess.Contains(policy.EFFECT_DENY) {
		trc.Decision("[explicit deny] found in resource policy")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}

	// If same account, access is granted if the Principal has access
	if evalIsSameAccount(s.ac.Principal, s.ac.Resource) {
		if pAccess.Contains(policy.EFFECT_ALLOW) {
			trc.Decision("[allow] access granted via same-account identity policy")
			return &Result{Trace: trc, IsAllowed: true}, nil
		}

		// TODO(nsiow) implement this behavior
		// if evalSameAccountExplicitPrincipalCase(ac.Principal, ac.Resource) {
		// 	trc.Decision("[allow] access granted via same-account explicit-resource-policy case")
		// 	return &Result{Trace: trc, IsAllowed: true}, nil
		// }

		trc.Decision("[implicit deny] no identity-based policy allows this action")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}

	// If x-account, access is granted if the Principal has access and the Resource permits that
	// access
	if pAccess.Contains(policy.EFFECT_ALLOW) && rAccess.Contains(policy.EFFECT_ALLOW) {
		trc.Decision("[allow] access granted via x-account identity + resource policies")
		return &Result{Trace: trc, IsAllowed: true}, nil
	}
	if pAccess.Contains(policy.EFFECT_ALLOW) && !rAccess.Contains(policy.EFFECT_ALLOW) {
		trc.Decision("[implicit deny] x-account, missing resource policy access")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}
	if !pAccess.Contains(policy.EFFECT_ALLOW) && rAccess.Contains(policy.EFFECT_ALLOW) {
		trc.Decision("[implicit deny] x-account, missing identity policy access")
		return &Result{Trace: trc, IsAllowed: false}, nil
	}

	// We fell through and no access was granted from either side
	trc.Decision("[implicit deny] x-account, missing both identity + resource access")
	return &Result{Trace: trc, IsAllowed: false}, nil
}
