package entities

// UniverseBuilder provides convenience functions for constructing universes using the builder
// pattern
type UniverseBuilder struct {
	uv *Universe
}

// NewUniverse creates and returns a new, empty universe
func NewBuilder() *UniverseBuilder {
	return &UniverseBuilder{uv: NewUniverse()}
}

// WithAccounts adds the provided accounts to the universe under construction
func (b *UniverseBuilder) WithAccounts(accounts ...Account) *UniverseBuilder {
	for _, a := range accounts {
		b.uv.PutAccount(a)
	}
	return b
}

// WithGroups adds the provided groups to the universe under construction
func (b *UniverseBuilder) WithGroups(groups ...Group) *UniverseBuilder {
	for _, g := range groups {
		b.uv.PutGroup(g)
	}
	return b
}

// WithPolicies adds the provided policies to the universe under construction
func (b *UniverseBuilder) WithPolicies(policies ...ManagedPolicy) *UniverseBuilder {
	for _, p := range policies {
		b.uv.PutPolicy(p)
	}
	return b
}

// WithPrincipals adds the provided principals to the universe under construction
func (b *UniverseBuilder) WithPrincipals(principals ...Principal) *UniverseBuilder {
	for _, p := range principals {
		b.uv.PutPrincipal(p)
	}
	return b
}

// WithResources adds the provided resources to the universe under construction
func (b *UniverseBuilder) WithResources(resources ...Resource) *UniverseBuilder {
	for _, r := range resources {
		b.uv.PutResource(r)
	}
	return b
}

// Build returns the universe constructed from the With* invocations thus far
func (b *UniverseBuilder) Build() *Universe {
	return b.uv
}
