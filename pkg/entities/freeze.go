package entities

import (
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
)

// -------------------------------------------------------------------------------------------------
// Universe
//
// TODO(nsiow) does it make any sense for principals/resources to have universe pointers if we are
//             going to allow freezing from other universes
// TODO(nsiow) update these to not be a method and instead take a list of universes
// -------------------------------------------------------------------------------------------------

func (u *Universe) FrozenPrincipals(overlay *Universe) ([]FrozenPrincipal, error) {
	var fs []FrozenPrincipal

	uvs := u.Overlay(overlay)
	for _, uv := range uvs {
		for p := range uv.Principals() {
			f, err := p.FreezeWith(uvs...)
			if err != nil {
				return nil, err
			}
			fs = append(fs, f)
		}
	}

	return fs, nil
}

func (u *Universe) FrozenResources(overlay *Universe) ([]FrozenResource, error) {
	var fs []FrozenResource

	uvs := u.Overlay(overlay)
	for _, uv := range uvs {
		for r := range uv.Resources() {
			f, err := r.FreezeWith(uvs...)
			if err != nil {
				return nil, err
			}
			fs = append(fs, f)
		}
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
	OrgNodes []FrozenOrgNode
}

func (a *Account) Freeze() (FrozenAccount, error) {
	if a.uv == nil {
		return FrozenAccount{}, fmt.Errorf("cannot freeze; account is missing universe: %s", a.Id)
	}

	return a.FreezeWith(a.uv)
}

func (a *Account) FreezeWith(universes ...*Universe) (FrozenAccount, error) {
	frozen := FrozenAccount{
		Id:       a.Id,
		OrgId:    a.OrgId,
		OrgPaths: a.OrgPaths,
	}

	for _, node := range a.OrgNodes {
		frozenNode, err := freezeOrgNode(&node, universes...)
		if err != nil {
			return FrozenAccount{}, err
		}

		frozen.OrgNodes = append(frozen.OrgNodes, frozenNode)
	}

	return frozen, nil
}

// -------------------------------------------------------------------------------------------------
// OrgNode
// -------------------------------------------------------------------------------------------------

type FrozenOrgNode struct {
	Id   string
	Type string
	Arn  string
	Name string

	SCPs []ManagedPolicy
	RCPs []ManagedPolicy
}

func freezeOrgNode(node *OrgNode, universes ...*Universe) (FrozenOrgNode, error) {
	frozen := FrozenOrgNode{
		Id:   node.Id,
		Type: node.Type,
		Arn:  node.Arn,
		Name: node.Name,
	}

	for _, arn := range node.SCPs {
		policies, err := freezePolicy(arn, universes...)
		if err != nil {
			return FrozenOrgNode{}, err
		}
		frozen.SCPs = append(frozen.SCPs, policies)
	}

	for _, arn := range node.RCPs {
		policies, err := freezePolicy(arn, universes...)
		if err != nil {
			return FrozenOrgNode{}, err
		}
		frozen.RCPs = append(frozen.RCPs, policies)
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
		return FrozenGroup{}, fmt.Errorf("cannot freeze; group is missing universe: %s", g.Arn)
	}

	return g.FreezeWith(g.uv)
}

func (g *Group) FreezeWith(universes ...*Universe) (FrozenGroup, error) {
	f := FrozenGroup{
		Type:           g.Type,
		AccountId:      g.AccountId,
		Arn:            g.Arn,
		InlinePolicies: g.InlinePolicies,
	}

	var err error

	f.AttachedPolicies, err = freezePolicies(g.AttachedPolicies, universes...)
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
	Account            FrozenAccount `json:",omitzero"`
	AttachedPolicies   []ManagedPolicy
	Groups             []FrozenGroup `json:",omitzero"`
	PermissionBoundary ManagedPolicy `json:",omitzero"`
}

func (p *Principal) Freeze() (FrozenPrincipal, error) {
	if p.uv == nil {
		return FrozenPrincipal{}, fmt.Errorf("cannot freeze; principal is missing universe: %s", p.Arn)
	}

	return p.FreezeWith(p.uv)
}

func (p *Principal) FreezeWith(universes ...*Universe) (FrozenPrincipal, error) {
	f := FrozenPrincipal{
		Type:           p.Type,
		AccountId:      p.AccountId,
		Arn:            p.Arn,
		Tags:           p.Tags,
		InlinePolicies: p.InlinePolicies,
	}

	var err error

	for _, uv := range universes {
		if account, ok := uv.Account(f.AccountId); ok {
			f.Account, err = account.FreezeWith(uv)
			if err != nil {
				return FrozenPrincipal{}, err
			}
		}
	}

	if len(p.AttachedPolicies) > 0 {
		f.AttachedPolicies, err = freezePolicies(p.AttachedPolicies, universes...)
		if err != nil {
			return FrozenPrincipal{}, err
		}
	}

	if len(p.Groups) > 0 {
		f.Groups, err = freezeGroupsByArn(p.Groups, universes...)
		if err != nil {
			return FrozenPrincipal{}, err
		}
	}

	if len(p.PermissionsBoundary) > 0 {
		f.PermissionBoundary, err = freezePolicy(p.PermissionsBoundary, universes...)
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
	Tags      []Tag `json:",omitzero"`

	Policy  policy.Policy `json:",omitzero"`
	Account FrozenAccount `json:",omitzero"`
}

func (r *Resource) Freeze() (FrozenResource, error) {
	if r.uv == nil {
		return FrozenResource{},
			fmt.Errorf("cannot freeze; resource is missing universe: %s", r.Arn)
	}

	return r.FreezeWith(r.uv)
}

func (r *Resource) FreezeWith(universes ...*Universe) (FrozenResource, error) {
	f := FrozenResource{
		Type:      r.Type,
		AccountId: r.AccountId,
		Region:    r.Region,
		Arn:       r.Arn,
		Tags:      r.Tags,
		Policy:    r.Policy,
	}

	var err error

	for _, uv := range universes {
		if account, ok := uv.Account(f.AccountId); ok {
			f.Account, err = account.FreezeWith(universes...)
			if err != nil {
				return FrozenResource{}, err
			}
		}
	}

	return f, nil
}

// -------------------------------------------------------------------------------------------------
// Helper functions
// -------------------------------------------------------------------------------------------------

func freezePolicy(arn Arn, universes ...*Universe) (ManagedPolicy, error) {
	for _, uv := range universes {
		pol, ok := uv.Policy(arn)
		if ok {
			return *pol, nil
		}
	}

	return ManagedPolicy{}, fmt.Errorf("cannot find policy with arn: %s", arn)
}

func freezePolicies(arns []Arn, universes ...*Universe) ([]ManagedPolicy, error) {
	policies := make([]ManagedPolicy, len(arns))

	for i, arn := range arns {
		pol, err := freezePolicy(arn, universes...)
		if err != nil {
			return nil, err
		}

		policies[i] = pol
	}

	return policies, nil
}

func freezeGroupByArn(arn Arn, universes ...*Universe) (FrozenGroup, error) {
	for _, uv := range universes {
		grp, ok := uv.Group(arn)
		if ok {
			frozen, err := grp.FreezeWith(universes...)
			if err != nil {
				return FrozenGroup{}, err
			}
			return frozen, nil
		}
	}

	return FrozenGroup{}, fmt.Errorf("cannot find group with arn: %s", arn)

}

func freezeGroupsByArn(arns []Arn, universes ...*Universe) ([]FrozenGroup, error) {
	groups := make([]FrozenGroup, len(arns))

	for i, arn := range arns {
		grp, err := freezeGroupByArn(arn, universes...)
		if err != nil {
			return nil, err
		}

		groups[i] = grp
	}

	return groups, nil
}
