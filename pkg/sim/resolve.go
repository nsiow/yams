package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
)

// -------------------------------------------------------------------------------------------------
// Account
// -------------------------------------------------------------------------------------------------

type resolvedAccount struct {
	entities.Account

	ResolvedSCPs [][]entities.ManagedPolicy
}

func resolveAccount(id string, uv *entities.Universe) (resolvedAccount, error) {
	a, ok := uv.Account(id)
	if !ok {
		return resolvedAccount{}, nil
	}

	r := resolvedAccount{
		Account:      *a,
		ResolvedSCPs: make([][]entities.ManagedPolicy, len(a.SCPs)),
	}

	for i, layer := range a.SCPs {
		policies, err := resolvePolicies(layer, uv)
		if err != nil {
			return resolvedAccount{}, err
		}
		r.ResolvedSCPs[i] = policies
	}

	return r, nil
}

// -------------------------------------------------------------------------------------------------
// Group
// -------------------------------------------------------------------------------------------------

type resolvedGroup struct {
	entities.Group

	ResolvedPolicies []entities.ManagedPolicy
}

func resolveGroup(arn entities.Arn, uv *entities.Universe) (*resolvedGroup, error) {
	g, ok := uv.Group(arn)
	if !ok {
		return nil, fmt.Errorf("cannot find group with arn: %s", arn.String())
	}

	r := resolvedGroup{
		Group: *g,
	}

	var err error

	r.ResolvedPolicies, err = resolvePolicies(r.Policies, uv)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func resolveGroups(arns []entities.Arn, uv *entities.Universe) ([]resolvedGroup, error) {
	groups := make([]resolvedGroup, len(arns))

	for i, arn := range arns {
		grp, err := resolveGroup(arn, uv)
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

	ResolvedAccount            resolvedAccount
	ResolvedAttachedPolicies   []entities.ManagedPolicy
	ResolvedGroups             []resolvedGroup
	ResolvedPermissionBoundary entities.ManagedPolicy
}

func resolvePrincipal(arn entities.Arn, uv *entities.Universe) (*resolvedPrincipal, error) {
	p, ok := uv.Principal(arn)
	if !ok {
		return nil, fmt.Errorf("cannot find principal with arn: %s", arn.String())
	}

	r := resolvedPrincipal{
		Principal: *p,
	}

	var err error

	r.ResolvedAccount, err = resolveAccount(r.AccountId, uv)
	if err != nil {
		return nil, err
	}

	r.ResolvedAttachedPolicies, err = resolvePolicies(r.AttachedPolicies, uv)
	if err != nil {
		return nil, err
	}

	r.ResolvedGroups, err = resolveGroups(r.Groups, uv)
	if err != nil {
		return nil, err
	}

	if !r.PermissionsBoundary.Empty() {
		r.ResolvedPermissionBoundary, err = resolvePolicy(r.PermissionsBoundary, uv)
		if err != nil {
			return nil, err
		}
	}

	return &r, nil
}

func resolvePrincipals(uv *entities.Universe) ([]resolvedPrincipal, error) {
	resolved := make([]resolvedPrincipal, 0)

	for p := range uv.Principals() {
		r, err := resolvePrincipal(p.Arn, uv)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, *r)
	}

	return resolved, nil
}

// -------------------------------------------------------------------------------------------------
// Resource
// -------------------------------------------------------------------------------------------------

type resolvedResource struct {
	entities.Resource

	// TODO(nsiow) RCPs go here
	ResolvedAccount resolvedAccount
}

func resolveResource(arn entities.Arn, uv *entities.Universe) (*resolvedResource, error) {
	resource, ok := uv.Resource(arn)
	if !ok {
		return nil, fmt.Errorf("cannot find resource with arn: %s", arn.String())
	}

	r := resolvedResource{
		Resource: *resource,
	}

	var err error

	if uv.HasAccount(r.AccountId) {
		r.ResolvedAccount, err = resolveAccount(r.AccountId, uv)
		if err != nil {
			return nil, err
		}
	}

	return &r, nil
}

func resolveResources(uv *entities.Universe) ([]resolvedResource, error) {
	resolved := make([]resolvedResource, 0)

	for r := range uv.Resources() {
		r2, err := resolveResource(r.Arn, uv)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, *r2)
	}

	return resolved, nil
}

// -------------------------------------------------------------------------------------------------
// Policies
// -------------------------------------------------------------------------------------------------

func resolvePolicy(arn entities.Arn, uv *entities.Universe) (entities.ManagedPolicy, error) {
	pol, ok := uv.Policy(arn)
	if !ok {
		return entities.ManagedPolicy{}, fmt.Errorf("cannot find policy with arn: %s", arn.String())
	}

	return *pol, nil
}

// FIXME(nsiow) change all mentions of "universe" to "uv" for conciseness
func resolvePolicies(arns []entities.Arn, uv *entities.Universe) ([]entities.ManagedPolicy, error) {
	policies := make([]entities.ManagedPolicy, len(arns))

	for i, arn := range arns {
		pol, err := resolvePolicy(arn, uv)
		if err != nil {
			return nil, err
		}

		policies[i] = pol
	}

	return policies, nil
}
