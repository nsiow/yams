package entities

import (
	"crypto/sha1"
	"encoding/hex"
	"slices"
	"time"
)

// Overlay represents a named collection of entity overrides that can be layered
// on top of a base Universe for simulation. Overlays act as a read-through cache,
// where lookups first check the overlay before falling back to the base universe.
type Overlay struct {
	// Name is the human-readable name of the overlay
	Name string

	// ID is the unique identifier derived from sha1 hash of the name
	ID string

	// CreatedAt is the timestamp when the overlay was created
	CreatedAt time.Time

	// Universe contains the overlay's entity collections
	Universe *Universe
}

// NewOverlay creates a new overlay with the given name. The ID is automatically
// generated from the sha1 hash of the name.
func NewOverlay(name string) *Overlay {
	return &Overlay{
		Name:      name,
		ID:        GenerateOverlayID(name),
		CreatedAt: time.Now(),
		Universe:  NewUniverse(),
	}
}

// NewOverlayWithTime creates a new overlay with a specific creation time.
// Useful for deserialization or testing.
func NewOverlayWithTime(name string, createdAt time.Time) *Overlay {
	return &Overlay{
		Name:      name,
		ID:        GenerateOverlayID(name),
		CreatedAt: createdAt,
		Universe:  NewUniverse(),
	}
}

// GenerateOverlayID creates a deterministic ID from the overlay name using sha1.
func GenerateOverlayID(name string) string {
	h := sha1.New()
	h.Write([]byte(name))
	return hex.EncodeToString(h.Sum(nil))
}

// Size returns the total number of entities in the overlay
func (o *Overlay) Size() int {
	if o.Universe == nil {
		return 0
	}
	return o.Universe.Size()
}

// IsEmpty returns true if the overlay contains no entities
func (o *Overlay) IsEmpty() bool {
	return o.Size() == 0
}

// NumPrincipals returns the number of principals in the overlay
func (o *Overlay) NumPrincipals() int {
	if o.Universe == nil {
		return 0
	}
	return o.Universe.NumPrincipals()
}

// NumResources returns the number of resources in the overlay
func (o *Overlay) NumResources() int {
	if o.Universe == nil {
		return 0
	}
	return o.Universe.NumResources()
}

// NumPolicies returns the number of policies in the overlay
func (o *Overlay) NumPolicies() int {
	if o.Universe == nil {
		return 0
	}
	return o.Universe.NumPolicies()
}

// NumAccounts returns the number of accounts in the overlay
func (o *Overlay) NumAccounts() int {
	if o.Universe == nil {
		return 0
	}
	return o.Universe.NumAccounts()
}

// NumGroups returns the number of groups in the overlay
func (o *Overlay) NumGroups() int {
	if o.Universe == nil {
		return 0
	}
	return o.Universe.NumGroups()
}

// OverlayData is a JSON-serializable representation of an Overlay.
// It contains all entity data in a flat, exportable format.
type OverlayData struct {
	Name       string          `json:"name"`
	ID         string          `json:"id"`
	CreatedAt  time.Time       `json:"createdAt"`
	Accounts   []Account       `json:"accounts,omitempty"`
	Groups     []Group         `json:"groups,omitempty"`
	Policies   []ManagedPolicy `json:"policies,omitempty"`
	Principals []Principal     `json:"principals,omitempty"`
	Resources  []Resource      `json:"resources,omitempty"`
}

// ToData converts an Overlay to its serializable representation.
func (o *Overlay) ToData() OverlayData {
	data := OverlayData{
		Name:      o.Name,
		ID:        o.ID,
		CreatedAt: o.CreatedAt,
	}

	if o.Universe == nil {
		return data
	}

	// Collect accounts
	for a := range o.Universe.Accounts() {
		data.Accounts = append(data.Accounts, *a)
	}

	// Collect groups
	for g := range o.Universe.Groups() {
		data.Groups = append(data.Groups, *g)
	}

	// Collect policies
	for p := range o.Universe.Policies() {
		data.Policies = append(data.Policies, *p)
	}

	// Collect principals
	for p := range o.Universe.Principals() {
		data.Principals = append(data.Principals, *p)
	}

	// Collect resources
	for r := range o.Universe.Resources() {
		data.Resources = append(data.Resources, *r)
	}

	return data
}

// FromData creates an Overlay from its serializable representation.
func FromData(data OverlayData) *Overlay {
	o := &Overlay{
		Name:      data.Name,
		ID:        data.ID,
		CreatedAt: data.CreatedAt,
		Universe:  NewUniverse(),
	}

	// Load accounts
	for _, a := range data.Accounts {
		o.Universe.PutAccount(a)
	}

	// Load groups
	for _, g := range data.Groups {
		o.Universe.PutGroup(g)
	}

	// Load policies
	for _, p := range data.Policies {
		o.Universe.PutPolicy(p)
	}

	// Load principals
	for _, p := range data.Principals {
		o.Universe.PutPrincipal(p)
	}

	// Load resources
	for _, r := range data.Resources {
		o.Universe.PutResource(r)
	}

	return o
}

// OverlaySummary is a lightweight representation of an Overlay for listing.
type OverlaySummary struct {
	Name          string    `json:"name"`
	ID            string    `json:"id"`
	CreatedAt     time.Time `json:"createdAt"`
	NumPrincipals int       `json:"numPrincipals"`
	NumResources  int       `json:"numResources"`
	NumPolicies   int       `json:"numPolicies"`
	NumAccounts   int       `json:"numAccounts"`
	NumGroups     int       `json:"numGroups"`
}

// Summary returns a lightweight summary of the overlay.
func (o *Overlay) Summary() OverlaySummary {
	return OverlaySummary{
		Name:          o.Name,
		ID:            o.ID,
		CreatedAt:     o.CreatedAt,
		NumPrincipals: o.NumPrincipals(),
		NumResources:  o.NumResources(),
		NumPolicies:   o.NumPolicies(),
		NumAccounts:   o.NumAccounts(),
		NumGroups:     o.NumGroups(),
	}
}

// Clone creates a deep copy of the overlay with a new name and ID.
func (o *Overlay) Clone(newName string) *Overlay {
	clone := NewOverlay(newName)

	if o.Universe != nil {
		// Copy all entities to the new universe
		for a := range o.Universe.Accounts() {
			clone.Universe.PutAccount(*a)
		}
		for g := range o.Universe.Groups() {
			clone.Universe.PutGroup(*g)
		}
		for p := range o.Universe.Policies() {
			clone.Universe.PutPolicy(*p)
		}
		for p := range o.Universe.Principals() {
			clone.Universe.PutPrincipal(*p)
		}
		for r := range o.Universe.Resources() {
			clone.Universe.PutResource(*r)
		}
	}

	return clone
}

// PrincipalArns returns a sorted list of all principal ARNs in the overlay.
func (o *Overlay) PrincipalArns() []string {
	if o.Universe == nil {
		return nil
	}
	arns := o.Universe.PrincipalArns()
	slices.Sort(arns)
	return arns
}

// ResourceArns returns a sorted list of all resource ARNs in the overlay.
func (o *Overlay) ResourceArns() []string {
	if o.Universe == nil {
		return nil
	}
	arns := o.Universe.ResourceArns()
	slices.Sort(arns)
	return arns
}

// PolicyArns returns a sorted list of all policy ARNs in the overlay.
func (o *Overlay) PolicyArns() []string {
	if o.Universe == nil {
		return nil
	}
	arns := o.Universe.PolicyArns()
	slices.Sort(arns)
	return arns
}

// GroupArns returns a sorted list of all group ARNs in the overlay.
func (o *Overlay) GroupArns() []string {
	if o.Universe == nil {
		return nil
	}
	arns := o.Universe.GroupArns()
	slices.Sort(arns)
	return arns
}

// AccountIds returns a sorted list of all account IDs in the overlay.
func (o *Overlay) AccountIds() []string {
	if o.Universe == nil {
		return nil
	}
	var ids []string
	for a := range o.Universe.Accounts() {
		ids = append(ids, a.Id)
	}
	slices.Sort(ids)
	return ids
}
