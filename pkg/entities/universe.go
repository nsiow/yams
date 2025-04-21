package entities

import (
	"iter"
	"maps"
)

// -------------------------------------------------------------------------------------------------
// Universe
// -------------------------------------------------------------------------------------------------

// Universe contains the definition of all accounts/policies/principals/resources used for
// simulation.
//
// In order for something to be considered for simulation, it must be a part of this universe
type Universe struct {
	accounts   map[string]Account
	groups     map[Arn]Group
	policies   map[Arn]ManagedPolicy
	principals map[Arn]Principal
	resources  map[Arn]Resource
}

// NewUniverse creates and returns a new, empty universe
func NewUniverse() *Universe {
	return &Universe{
		accounts:   make(map[string]Account),
		groups:     make(map[Arn]Group),
		policies:   make(map[Arn]ManagedPolicy),
		principals: make(map[Arn]Principal),
		resources:  make(map[Arn]Resource),
	}
}

// -------------------------------------------------------------------------------------------------
// Accounts
// -------------------------------------------------------------------------------------------------

// Accounts returns an iterator over all the Account entities known to the universe
func (u *Universe) Accounts() iter.Seq[Account] {
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
	return &a, ok
}

// PutAccount saves the provided account into the universe, updating the definition if needed
func (u *Universe) PutAccount(a Account) {
	a.uv = u
	u.accounts[a.Id] = a
}

// RemoveAccount removes the account referenced by the provided id
func (u *Universe) RemoveAccount(id string) {
	delete(u.accounts, id)
}

// -------------------------------------------------------------------------------------------------
// Groups
// -------------------------------------------------------------------------------------------------

// Groups returns an iterator over all the Group entities known to the universe
func (u *Universe) Groups() iter.Seq[Group] {
	return maps.Values(u.groups)
}

// HasGroup returns whether or not the specified group exists in the universe
func (u *Universe) HasGroup(arn Arn) bool {
	_, ok := u.Group(arn)
	return ok
}

// Group attempts to retrieve the group based on its ARN
func (u *Universe) Group(arn Arn) (*Group, bool) {
	g, ok := u.groups[arn]
	return &g, ok
}

// PutGroup saves the provided group into the universe, updating the definition if needed
func (u *Universe) PutGroup(g Group) {
	g.uv = u
	u.groups[g.Arn] = g
}

// RemoveGroup removes the group referenced by the provided ARN
func (u *Universe) RemoveGroup(arn Arn) {
	delete(u.groups, arn)
}

// -------------------------------------------------------------------------------------------------
// Policies
// -------------------------------------------------------------------------------------------------

// Policies returns an iterator over all the IAM policies known to the universe
//
// This includes any policy with an ARN, e.g. managed policies, SCPs, etc. It does not include
// inline Principal or Resource policies
func (u *Universe) Policies() iter.Seq[ManagedPolicy] {
	return maps.Values(u.policies)
}

// HasPolicy returns whether or not the specified policy exists in the universe
func (u *Universe) HasPolicy(arn Arn) bool {
	_, ok := u.Policy(arn)
	return ok
}

// Policy attempts to retrieve the policy based on its ARN
func (u *Universe) Policy(arn Arn) (*ManagedPolicy, bool) {
	p, ok := u.policies[arn]
	return &p, ok
}

// PutPolicy saves the provided policy into the universe, updating the definition if needed
func (u *Universe) PutPolicy(p ManagedPolicy) {
	u.policies[p.Arn] = p
}

// RemovePolicy removes the policy referenced by the provided ARN
func (u *Universe) RemovePolicy(arn Arn) {
	delete(u.policies, arn)
}

// -------------------------------------------------------------------------------------------------
// Principles
// -------------------------------------------------------------------------------------------------

// Principals returns an iterator over all the Principal entities known to the universe
func (u *Universe) Principals() iter.Seq[Principal] {
	return maps.Values(u.principals)
}

// HasPrincipal returns whether or not the specified principal exists in the universe
func (u *Universe) HasPrincipal(arn Arn) bool {
	_, ok := u.Principal(arn)
	return ok
}

// Principal attempts to retrieve the principal based on its ARN
func (u *Universe) Principal(arn Arn) (*Principal, bool) {
	p, ok := u.principals[arn]
	return &p, ok
}

// PutPrincipal saves the provided principal into the universe, updating the definition if needed
func (u *Universe) PutPrincipal(p Principal) {
	p.uv = u
	u.principals[p.Arn] = p
	// TODO(nsiow) should this also update the resources where relevant (user/role)?
}

// RemovePrincipal removes the principal referenced by the provided ARN
func (u *Universe) RemovePrincipal(arn Arn) {
	delete(u.principals, arn)
}

// -------------------------------------------------------------------------------------------------
// Resources
// -------------------------------------------------------------------------------------------------

// Resources returns an iterator over all the Resource entities known to the universe
func (u *Universe) Resources() iter.Seq[Resource] {
	return maps.Values(u.resources)
}

// HasResource returns whether or not the specified resource exists in the universe
func (u *Universe) HasResource(arn Arn) bool {
	_, ok := u.Resource(arn)
	return ok
}

// Resource attempts to retrieve the resource based on its ARN
func (u *Universe) Resource(arn Arn) (*Resource, bool) {
	r, ok := u.resources[arn]
	return &r, ok
}

// PutResource saves the provided resource into the universe, updating the definition if needed
func (u *Universe) PutResource(r Resource) {
	r.uv = u
	u.resources[r.Arn] = r
}

// RemoveResource removes the resource referenced by the provided ARN
func (u *Universe) RemoveResource(arn Arn) {
	delete(u.resources, arn)
}
