package entities

import (
	"fmt"
)

// -------------------------------------------------------------------------------------------------
// Universe
// -------------------------------------------------------------------------------------------------

func (u *Universe) FrozenPrincipals() ([]FrozenPrincipal, error) {
	var fs []FrozenPrincipal

	for p := range u.Principals() {
		f, err := p.Freeze()
		if err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}

	return fs, nil
}

func (u *Universe) FrozenResources() ([]FrozenResource, error) {
	var fs []FrozenResource

	for r := range u.Resources() {
		f, err := r.Freeze()
		if err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}

	return fs, nil
}

// -------------------------------------------------------------------------------------------------
// Account
// -------------------------------------------------------------------------------------------------

type FrozenAccount struct {
	Account
	FrozenSCPs [][]ManagedPolicy
}

func (a *Account) Freeze() (FrozenAccount, error) {
	if a.uv == nil {
		return FrozenAccount{}, fmt.Errorf("cannot freeze; account is missing universe: %s", a.Id)
	}

	frozen := FrozenAccount{
		Account: *a,
	}

	for _, layer := range a.SCPs {
		policies, err := freezePolicies(layer, a.uv)
		if err != nil {
			return FrozenAccount{}, err
		}
		frozen.FrozenSCPs = append(frozen.FrozenSCPs, policies)
	}

	return frozen, nil
}

// -------------------------------------------------------------------------------------------------
// Group
// -------------------------------------------------------------------------------------------------

type FrozenGroup struct {
	Group
	FrozenPolicies []ManagedPolicy
}

func (g *Group) Freeze() (FrozenGroup, error) {
	if g.uv == nil {
		return FrozenGroup{}, fmt.Errorf("cannot freeze; group is missing universe: %s", g.Arn.String())
	}

	f := FrozenGroup{
		Group: *g,
	}

	var err error

	f.FrozenPolicies, err = freezePolicies(f.Policies, g.uv)
	if err != nil {
		return FrozenGroup{}, err
	}

	return f, nil
}

// -------------------------------------------------------------------------------------------------
// Principal
// -------------------------------------------------------------------------------------------------

type FrozenPrincipal struct {
	Principal
	FrozenAccount            FrozenAccount
	FrozenAttachedPolicies   []ManagedPolicy
	FrozenGroups             []FrozenGroup
	FrozenPermissionBoundary ManagedPolicy
}

func (p *Principal) Freeze() (FrozenPrincipal, error) {
	if p.uv == nil {
		return FrozenPrincipal{},
			fmt.Errorf("cannot freeze; principal is missing universe: %s", p.Arn.String())
	}

	f := FrozenPrincipal{
		Principal: *p,
	}

	var err error

	if account, ok := f.uv.Account(f.AccountId); ok {
		f.FrozenAccount, err = account.Freeze()
		if err != nil {
			return FrozenPrincipal{}, err
		}
	}

	if len(f.AttachedPolicies) > 0 {
		f.FrozenAttachedPolicies, err = freezePolicies(f.AttachedPolicies, f.uv)
		if err != nil {
			return FrozenPrincipal{}, err
		}
	}

	if len(f.Groups) > 0 {
		f.FrozenGroups, err = freezeGroupsByArn(f.Groups, f.uv)
		if err != nil {
			return FrozenPrincipal{}, err
		}
	}

	if !f.PermissionsBoundary.Empty() {
		f.FrozenPermissionBoundary, err = freezePolicy(f.PermissionsBoundary, f.uv)
		if err != nil {
			return FrozenPrincipal{}, err
		}
	}

	return f, nil
}

// -------------------------------------------------------------------------------------------------
// Resource
// -------------------------------------------------------------------------------------------------

type FrozenResource struct {
	Resource
	// TODO(nsiow) RCPs go here
	FrozenAccount FrozenAccount
}

func (r *Resource) Freeze() (FrozenResource, error) {
	if r.uv == nil {
		return FrozenResource{},
			fmt.Errorf("cannot freeze; resource is missing universe: %s", r.Arn.String())
	}

	f := FrozenResource{
		Resource: *r,
	}

	var err error

	if account, ok := f.uv.Account(f.AccountId); ok {
		f.FrozenAccount, err = account.Freeze()
		if err != nil {
			return FrozenResource{}, err
		}
	}

	return f, nil
}

// -------------------------------------------------------------------------------------------------
// Helper functions
// -------------------------------------------------------------------------------------------------

func freezePolicy(arn Arn, uv *Universe) (ManagedPolicy, error) {
	pol, ok := uv.Policy(arn)
	if !ok {
		return ManagedPolicy{}, fmt.Errorf("cannot find policy with arn: %s", arn.String())
	}

	return *pol, nil
}

// FIXME(nsiow) change all mentions of "universe" to "uv" for conciseness
func freezePolicies(arns []Arn, uv *Universe) ([]ManagedPolicy, error) {
	policies := make([]ManagedPolicy, len(arns))

	for i, arn := range arns {
		pol, err := freezePolicy(arn, uv)
		if err != nil {
			return nil, err
		}

		policies[i] = pol
	}

	return policies, nil
}

func freezeGroupByArn(arn Arn, uv *Universe) (FrozenGroup, error) {
	grp, ok := uv.Group(arn)
	if !ok {
		return FrozenGroup{}, fmt.Errorf("cannot find group with arn: %s", arn.String())
	}

	frozen, err := grp.Freeze()
	if err != nil {
		return FrozenGroup{}, err
	}
	return frozen, nil
}

func freezeGroupsByArn(arns []Arn, uv *Universe) ([]FrozenGroup, error) {
	groups := make([]FrozenGroup, len(arns))

	for i, arn := range arns {
		grp, err := freezeGroupByArn(arn, uv)
		if err != nil {
			return nil, err
		}

		groups[i] = grp
	}

	return groups, nil
}
