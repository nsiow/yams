package entities

// An Universe corresponds to a set of coexistant Principals and Resources
type Universe struct {
	Principals []Principal
	Resources  []Resource
}
