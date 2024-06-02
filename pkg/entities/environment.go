package entities

// An Environment corresponds to a set of coexistant Principals and Resources
type Environment struct {
	Principals []Principal
	Resources  []Resource
}
