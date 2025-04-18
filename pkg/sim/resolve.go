package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// -------------------------------------------------------------------------------------------------
// Account
// -------------------------------------------------------------------------------------------------

type resolvedAccount struct {
	entities.Account

	ResolvedSCPs [][]entities.Policy
}

func resolveAccount(accountId string, universe *entities.Universe) (*resolvedAccount, error) {
	a, ok := universe.Account(accountId)
	if !ok {
		return nil, fmt.Errorf("unable to locate account with arn: %s", accountId)
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

	ResolvedAccount            resolvedAccount
	ResolvedAttachedPolicies   []entities.Policy
	ResolvedGroups             []resolvedGroup
	ResolvedPermissionBoundary entities.Policy
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

	r.ResolvedAttachedPolicies, err = resolvePolicies(r.AttachedPolicies, universe)
	if err != nil {
		return nil, err
	}

	r.ResolvedGroups, err = resolveGroups(r.Groups, universe)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// -------------------------------------------------------------------------------------------------
// Resource
// -------------------------------------------------------------------------------------------------

type resolvedResource struct {
	entities.Resource

	ResolvedPolicy policy.Policy
}

// -------------------------------------------------------------------------------------------------
// Policies
// -------------------------------------------------------------------------------------------------

func resolvePolicy(arn entities.Arn, universe *entities.Universe) (*entities.Policy, error) {
	pol, ok := universe.Policy(arn)
	if !ok {
		return nil, fmt.Errorf("unable to locate policy with arn: %s", arn.String())
	}

	return pol, nil
}

func resolvePolicies(arns []entities.Arn, universe *entities.Universe) ([]entities.Policy, error) {
	policies := make([]entities.Policy, len(arns))

	for i, arn := range arns {
		pol, err := resolvePolicy(arn, universe)
		if err != nil {
			return nil, err
		}

		policies[i] = *pol
	}

	return policies, nil
}
