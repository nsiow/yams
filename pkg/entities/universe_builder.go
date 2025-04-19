package entities

// UniverseBuilder provides convenience functions for constructing universes using the builder
// pattern
type UniverseBuilder struct {
	universe *Universe
}

// NewUniverse creates and returns a new, empty universe
func NewBuilder() *UniverseBuilder {
	return &UniverseBuilder{universe: NewUniverse()}
}

// WithAccounts adds the provided accounts to the universe under construction
func (ub *UniverseBuilder) WithAccounts(accounts ...Account) *UniverseBuilder {
	for _, a := range accounts {
		ub.universe.PutAccount(a)
	}
	return ub
}

// WithGroups adds the provided groups to the universe under construction
func (ub *UniverseBuilder) WithGroups(groups ...Group) *UniverseBuilder {
	for _, g := range groups {
		ub.universe.PutGroup(g)
	}
	return ub
}

// WithPolicies adds the provided policies to the universe under construction
func (ub *UniverseBuilder) WithPolicies(policies ...ManagedPolicy) *UniverseBuilder {
	for _, p := range policies {
		ub.universe.PutPolicy(p)
	}
	return ub
}

// WithPrincipals adds the provided principals to the universe under construction
func (ub *UniverseBuilder) WithPrincipals(principals ...Principal) *UniverseBuilder {
	for _, p := range principals {
		ub.universe.PutPrincipal(p)
	}
	return ub
}

// WithResources adds the provided resources to the universe under construction
func (ub *UniverseBuilder) WithResources(resources ...Resource) *UniverseBuilder {
	for _, r := range resources {
		ub.universe.PutResource(r)
	}
	return ub
}

// Build returns the universe constructed from the With* invocations thus far
func (ub *UniverseBuilder) Build() *Universe {
	return ub.universe
}
