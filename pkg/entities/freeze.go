package entities

import (
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
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
	Id       string
	OrgId    string
	OrgPaths []string

	SCPs [][]ManagedPolicy
}

func (a *Account) Freeze() (FrozenAccount, error) {
	if a.uv == nil {
		return FrozenAccount{}, fmt.Errorf("cannot freeze; account is missing uv: %s", a.Id)
	}

	frozen := FrozenAccount{
		Id:       a.Id,
		OrgId:    a.OrgId,
		OrgPaths: a.OrgPaths,
	}

	for _, layer := range a.SCPs {
		policies, err := freezePolicies(layer, a.uv)
		if err != nil {
			return FrozenAccount{}, err
		}
		frozen.SCPs = append(frozen.SCPs, policies)
	}

	return frozen, nil
}

// -------------------------------------------------------------------------------------------------
// Group
// -------------------------------------------------------------------------------------------------

type FrozenGroup struct {
	Type      string
	AccountId string
	Arn       Arn

	InlinePolicies   []policy.Policy
	AttachedPolicies []ManagedPolicy
}

func (g *Group) Freeze() (FrozenGroup, error) {
	if g.uv == nil {
		return FrozenGroup{}, fmt.Errorf("cannot freeze; group is missing uv: %s", g.Arn.String())
	}

	f := FrozenGroup{
		Type:           g.Type,
		AccountId:      g.AccountId,
		Arn:            g.Arn,
		InlinePolicies: g.InlinePolicies,
	}

	var err error

	f.AttachedPolicies, err = freezePolicies(g.AttachedPolicies, g.uv)
	if err != nil {
		return FrozenGroup{}, err
	}

	return f, nil
}

// -------------------------------------------------------------------------------------------------
// Principal
// -------------------------------------------------------------------------------------------------

type FrozenPrincipal struct {
	Type      string
	AccountId string
	Arn       Arn
	Tags      []Tag

	InlinePolicies     []policy.Policy
	Account            FrozenAccount
	AttachedPolicies   []ManagedPolicy
	Groups             []FrozenGroup
	PermissionBoundary ManagedPolicy
}

func (p *Principal) Freeze() (FrozenPrincipal, error) {
	if p.uv == nil {
		return FrozenPrincipal{},
			fmt.Errorf("cannot freeze; principal is missing uv: %s", p.Arn.String())
	}

	f := FrozenPrincipal{
		Type:           p.Type,
		AccountId:      p.AccountId,
		Arn:            p.Arn,
		Tags:           p.Tags,
		InlinePolicies: p.InlinePolicies,
	}

	var err error

	if account, ok := p.uv.Account(f.AccountId); ok {
		f.Account, err = account.Freeze()
		if err != nil {
			return FrozenPrincipal{}, err
		}
	}

	if len(p.AttachedPolicies) > 0 {
		f.AttachedPolicies, err = freezePolicies(p.AttachedPolicies, p.uv)
		if err != nil {
			return FrozenPrincipal{}, err
		}
	}

	if len(p.Groups) > 0 {
		f.Groups, err = freezeGroupsByArn(p.Groups, p.uv)
		if err != nil {
			return FrozenPrincipal{}, err
		}
	}

	if !p.PermissionsBoundary.Empty() {
		f.PermissionBoundary, err = freezePolicy(p.PermissionsBoundary, p.uv)
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
	Type      string
	AccountId string
	Region    string
	Arn       Arn
	Tags      []Tag `json:"omitzero"`

	Policy policy.Policy
	// TODO(nsiow) RCPs go here
	Account FrozenAccount
}

func (r *Resource) Freeze() (FrozenResource, error) {
	if r.uv == nil {
		return FrozenResource{},
			fmt.Errorf("cannot freeze; resource is missing uv: %s", r.Arn.String())
	}

	f := FrozenResource{
		Type:      r.Type,
		AccountId: r.AccountId,
		Arn:       r.Arn,
		Tags:      r.Tags,
		Policy:    r.Policy,
	}

	var err error

	if account, ok := r.uv.Account(f.AccountId); ok {
		f.Account, err = account.Freeze()
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
