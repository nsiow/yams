package entities

import (
	"iter"
	"maps"
	"path"
	"strings"
	"sync"

	"github.com/nsiow/yams/internal/assets"
)

// -------------------------------------------------------------------------------------------------
// Universe
// -------------------------------------------------------------------------------------------------

type Arn = string

// Universe contains the definition of all accounts/policies/principals/resources used for
// simulation.
//
// In order for something to be considered for simulation, it must be a part of this universe
type Universe struct {
	mut        sync.RWMutex
	accounts   map[string]*Account
	groups     map[Arn]*Group
	policies   map[Arn]*ManagedPolicy
	principals map[Arn]*Principal
	resources  map[Arn]*Resource

	hasLoadedBasePolicies bool
}

// NewUniverse creates and returns a new, empty universe
func NewUniverse() *Universe {
	return &Universe{
		accounts:   make(map[string]*Account),
		groups:     make(map[Arn]*Group),
		policies:   make(map[Arn]*ManagedPolicy),
		principals: make(map[Arn]*Principal),
		resources:  make(map[Arn]*Resource),
	}
}

// Merge adds all entries in `other` [Universe] to this one
func (u *Universe) Merge(other *Universe) {
	other.mut.RLock()
	defer other.mut.RUnlock()

	for _, item := range other.accounts {
		u.PutAccount(*item)
	}
	for _, item := range other.groups {
		u.PutGroup(*item)
	}
	for _, item := range other.policies {
		u.PutPolicy(*item)
	}
	for _, item := range other.principals {
		u.PutPrincipal(*item)
	}
	for _, item := range other.resources {
		u.PutResource(*item)
	}
}

// Overlay returns a priority-order slice of the two [Universe]s combined
func (u *Universe) Overlay(other *Universe) []*Universe {
	if other == nil {
		return []*Universe{u}
	} else {
		return []*Universe{other, u}
	}
}

// Size returns the number of known entities in the universe
func (u *Universe) Size() int {
	u.mut.RLock()
	defer u.mut.RUnlock()

	return len(u.accounts) +
		len(u.groups) +
		len(u.policies) +
		len(u.principals) +
		len(u.resources)
}

// -------------------------------------------------------------------------------------------------
// Accounts
// -------------------------------------------------------------------------------------------------

// NumAccounts returns the number of accounts known to the universe
func (u *Universe) NumAccounts() int {
	u.mut.RLock()
	defer u.mut.RUnlock()
	return len(u.accounts)
}

// Accounts returns an iterator over all the Account entities known to the universe
func (u *Universe) Accounts() iter.Seq[*Account] {
	return maps.Values(u.accounts)
}

// HasAccount returns whether or not the specified account exists in the universe
func (u *Universe) HasAccount(id string) bool {
	_, ok := u.Account(id)
	return ok
}

// Account attempts to retrieve the account based on its id
func (u *Universe) Account(id string) (*Account, bool) {
	a, ok := u.accounts[id]
	return a, ok
}

// PutAccount saves the provided account into the universe, updating the definition if needed
func (u *Universe) PutAccount(a Account) {
	u.mut.Lock()
	defer u.mut.Unlock()

	a.uv = u
	u.accounts[a.Id] = &a
}

// RemoveAccount removes the account referenced by the provided id
func (u *Universe) RemoveAccount(id string) {
	u.mut.Lock()
	defer u.mut.Unlock()

	delete(u.accounts, id)
}

// -------------------------------------------------------------------------------------------------
// Groups
// -------------------------------------------------------------------------------------------------

// NumGroups returns the number of groups known to the universe
func (u *Universe) NumGroups() int {
	u.mut.RLock()
	defer u.mut.RUnlock()
	return len(u.groups)
}

// Groups returns an iterator over all the Group entities known to the universe
func (u *Universe) Groups() iter.Seq[*Group] {
	return maps.Values(u.groups)
}

// GroupArns returns a slice containing the ARNs of all known Groups
func (u *Universe) GroupArns() []string {
	u.mut.RLock()
	defer u.mut.RUnlock()

	arns := []string{}
	for _, g := range u.groups {
		arns = append(arns, g.Arn)
	}
	return arns
}

// HasGroup returns whether or not the specified group exists in the universe
func (u *Universe) HasGroup(arn Arn) bool {
	_, ok := u.Group(arn)
	return ok
}

// Group attempts to retrieve the group based on its ARN
func (u *Universe) Group(arn Arn) (*Group, bool) {
	g, ok := u.groups[arn]
	return g, ok
}

// PutGroup saves the provided group into the universe, updating the definition if needed
func (u *Universe) PutGroup(g Group) {
	u.mut.Lock()
	defer u.mut.Unlock()

	g.uv = u
	u.groups[g.Arn] = &g
}

// RemoveGroup removes the group referenced by the provided ARN
func (u *Universe) RemoveGroup(arn Arn) {
	u.mut.Lock()
	defer u.mut.Unlock()

	delete(u.groups, arn)
}

// -------------------------------------------------------------------------------------------------
// Policies
// -------------------------------------------------------------------------------------------------

// LoadBasePolicies bootstraps the Universe with base IAM policies that are present in every account
func (u *Universe) LoadBasePolicies() {
	if !u.hasLoadedBasePolicies {
		for arn, policy := range assets.ManagedPolicyData() {
			u.PutPolicy(ManagedPolicy{
				Type:      "AWS::IAM::Policy",
				AccountId: "AWS",
				Arn:       arn,
				Name:      path.Base(arn),
				Policy:    policy,
			})
		}

		u.hasLoadedBasePolicies = true
	}
}

// NumPolicies returns the number of policies known to the universe
func (u *Universe) NumPolicies() int {
	u.mut.RLock()
	defer u.mut.RUnlock()
	return len(u.policies)
}

// Policies returns an iterator over all the IAM policies known to the universe
//
// This includes any policy with an ARN, e.g. managed policies, SCPs, etc. It does not include
// inline Principal or Resource policies
func (u *Universe) Policies() iter.Seq[*ManagedPolicy] {
	return maps.Values(u.policies)
}

// PolicyArns returns a slice containing the ARNs of all known Policies
func (u *Universe) PolicyArns() []string {
	u.mut.RLock()
	defer u.mut.RUnlock()

	arns := []string{}
	for _, p := range u.policies {
		arns = append(arns, p.Arn)
	}
	return arns
}

// HasPolicy returns whether or not the specified policy exists in the universe
func (u *Universe) HasPolicy(arn Arn) bool {
	_, ok := u.Policy(arn)
	return ok
}

// Policy attempts to retrieve the policy based on its ARN
func (u *Universe) Policy(arn Arn) (*ManagedPolicy, bool) {
	p, ok := u.policies[arn]
	return p, ok
}

// PutPolicy saves the provided policy into the universe, updating the definition if needed
func (u *Universe) PutPolicy(p ManagedPolicy) {
	u.mut.Lock()
	defer u.mut.Unlock()

	u.policies[p.Arn] = &p
}

// RemovePolicy removes the policy referenced by the provided ARN
func (u *Universe) RemovePolicy(arn Arn) {
	u.mut.Lock()
	defer u.mut.Unlock()

	delete(u.policies, arn)
}

// -------------------------------------------------------------------------------------------------
// Principles
// -------------------------------------------------------------------------------------------------

// NumPrincipals returns the number of principals known to the universe
func (u *Universe) NumPrincipals() int {
	u.mut.RLock()
	defer u.mut.RUnlock()
	return len(u.principals)
}

// Principals returns an iterator over all the Principal entities known to the universe
func (u *Universe) Principals() iter.Seq[*Principal] {
	return maps.Values(u.principals)
}

// PrincipalArns returns a slice containing the ARNs of all known Principals
func (u *Universe) PrincipalArns() []string {
	u.mut.RLock()
	defer u.mut.RUnlock()

	arns := []string{}
	for _, p := range u.principals {
		arns = append(arns, p.Arn)
	}
	return arns
}

// HasPrincipal returns whether or not the specified principal exists in the universe
func (u *Universe) HasPrincipal(arn Arn) bool {
	_, ok := u.Principal(arn)
	return ok
}

// Principal attempts to retrieve the principal based on its ARN
func (u *Universe) Principal(arn Arn) (*Principal, bool) {
	p, ok := u.principals[arn]
	return p, ok
}

// PutPrincipal saves the provided principal into the universe, updating the definition if needed
func (u *Universe) PutPrincipal(p Principal) {
	u.mut.Lock()
	defer u.mut.Unlock()

	p.uv = u
	u.principals[p.Arn] = &p
	// TODO(nsiow) should this also update the resources where relevant (user/role)?
}

// RemovePrincipal removes the principal referenced by the provided ARN
func (u *Universe) RemovePrincipal(arn Arn) {
	u.mut.Lock()
	defer u.mut.Unlock()

	delete(u.principals, arn)
}

// -------------------------------------------------------------------------------------------------
// Resources
// -------------------------------------------------------------------------------------------------

// NumResources returns the number of resources known to the universe
func (u *Universe) NumResources() int {
	u.mut.RLock()
	defer u.mut.RUnlock()
	return len(u.resources)
}

func (u *Universe) subresource(arn Arn) (string, string) {
	// handle S3 objects
	if strings.HasPrefix(arn, "arn:aws:s3:::") && strings.Contains(arn, "/") {
		components := strings.SplitN(arn, "/", 2)
		return components[0], components[1]
	}

	return arn, ""
}

// Resources returns an iterator over all the Resource entities known to the universe
func (u *Universe) Resources() iter.Seq[*Resource] {
	return maps.Values(u.resources)
}

// ResourceArns returns a slice containing the ARNs of all known Resources
func (u *Universe) ResourceArns() []string {
	u.mut.RLock()
	defer u.mut.RUnlock()

	arns := []string{}
	for _, r := range u.resources {
		arns = append(arns, r.Arn)
	}
	return arns
}

// HasResource returns whether or not the specified resource exists in the universe
func (u *Universe) HasResource(arn Arn) bool {
	_, ok := u.Resource(arn)
	return ok
}

// Resource attempts to retrieve the resource based on its ARN
func (u *Universe) Resource(arn Arn) (*Resource, bool) {
	arn, path := u.subresource(arn)

	r, ok := u.resources[arn]
	if !ok {
		return nil, ok
	}

	if len(path) > 0 {
		subResource, err := r.SubResource(path)
		return subResource, err == nil
	}

	return r, ok
}

// PutResource saves the provided resource into the universe, updating the definition if needed
func (u *Universe) PutResource(r Resource) {
	u.mut.Lock()
	defer u.mut.Unlock()

	r.uv = u
	u.resources[r.Arn] = &r
}

// RemoveResource removes the resource referenced by the provided ARN
func (u *Universe) RemoveResource(arn Arn) {
	u.mut.Lock()
	defer u.mut.Unlock()

	delete(u.resources, arn)
}
