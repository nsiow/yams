package sim

import (
	"fmt"

	"github.com/nsiow/yams/internal/opt"
	"github.com/nsiow/yams/pkg/entities"
)

// -------------------------------------------------------------------------------------------------
// Account
// -------------------------------------------------------------------------------------------------

type resolvedAccount struct {
	entities.Account

	ResolvedSCPs [][]entities.Policy
}

func resolveAccount(id string, universe *entities.Universe) (*resolvedAccount, error) {
	a, ok := universe.Account(id)
	if !ok {
		return nil, nil
	}

	r := resolvedAccount{
		Account:      *a,
		ResolvedSCPs: make([][]entities.Policy, len(a.SCPs)),
	}

	for i, layer := range a.SCPs {
		policies, err := resolvePolicies(layer, universe)
		if err != nil {
			return nil, err
		}
		r.ResolvedSCPs[i] = policies
	}

	return &r, nil
}

// -------------------------------------------------------------------------------------------------
// Group
// -------------------------------------------------------------------------------------------------

type resolvedGroup struct {
	entities.Group

	ResolvedPolicies []entities.Policy
}

func resolveGroup(arn entities.Arn, universe *entities.Universe) (*resolvedGroup, error) {
	g, ok := universe.Group(arn)
	if !ok {
		return nil, fmt.Errorf("unable to locate group with arn: %s", arn.String())
	}

	r := resolvedGroup{
		Group: *g,
	}

	var err error

	r.ResolvedPolicies, err = resolvePolicies(r.Policies, universe)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func resolveGroups(arns []entities.Arn, universe *entities.Universe) ([]resolvedGroup, error) {
	groups := make([]resolvedGroup, len(arns))

	for i, arn := range arns {
		grp, err := resolveGroup(arn, universe)
		if err != nil {
			return nil, err
		}

		groups[i] = *grp
	}

	return groups, nil
}

// -------------------------------------------------------------------------------------------------
// Principal
// -------------------------------------------------------------------------------------------------

type resolvedPrincipal struct {
	entities.Principal

	ResolvedAccount            opt.Option[resolvedAccount]
	ResolvedAttachedPolicies   []entities.Policy
	ResolvedGroups             []resolvedGroup
	ResolvedPermissionBoundary opt.Option[entities.Policy]
}

func resolvePrincipal(arn entities.Arn, universe *entities.Universe) (*resolvedPrincipal, error) {
	p, ok := universe.Principal(arn)
	if !ok {
		return nil, fmt.Errorf("unable to locate principal with arn: %s", arn.String())
	}

	r := resolvedPrincipal{
		Principal: *p,
	}

	var err error

	if universe.HasAccount(r.AccountId) {
		resolved, err := resolveAccount(r.AccountId, universe)
		if err != nil {
			return nil, err
		}
		r.ResolvedAccount = opt.Some(*resolved)
	}

	r.ResolvedAttachedPolicies, err = resolvePolicies(r.AttachedPolicies, universe)
	if err != nil {
		return nil, err
	}

	r.ResolvedGroups, err = resolveGroups(r.Groups, universe)
	if err != nil {
		return nil, err
	}

	if !r.PermissionsBoundary.Empty() {
		resolved, err := resolvePolicy(r.PermissionsBoundary, universe)
		if err != nil {
			return nil, err
		}
		r.ResolvedPermissionBoundary = opt.Some(resolved)
	}

	return &r, nil
}

// -------------------------------------------------------------------------------------------------
// Resource
// -------------------------------------------------------------------------------------------------

type resolvedResource struct {
	entities.Resource

	// TODO(nsiow) RCPs go here
}

func resolveResource(arn entities.Arn, universe *entities.Universe) (*resolvedResource, error) {
	resource, ok := universe.Resource(arn)
	if !ok {
		return nil, fmt.Errorf("unable to locate resource with arn: %s", arn.String())
	}

	r := resolvedResource{
		Resource: *resource,
	}

	return &r, nil
}

// -------------------------------------------------------------------------------------------------
// Policies
// -------------------------------------------------------------------------------------------------

func resolvePolicy(arn entities.Arn, universe *entities.Universe) (entities.Policy, error) {
	pol, ok := universe.Policy(arn)
	if !ok {
		return entities.Policy{}, fmt.Errorf("unable to locate policy with arn: %s", arn.String())
	}

	return *pol, nil
}

func resolvePolicies(arns []entities.Arn, universe *entities.Universe) ([]entities.Policy, error) {
	policies := make([]entities.Policy, len(arns))

	for i, arn := range arns {
		pol, err := resolvePolicy(arn, universe)
		if err != nil {
			return nil, err
		}

		policies[i] = pol
	}

	return policies, nil
}
