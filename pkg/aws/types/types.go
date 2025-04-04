package types

// Service represents a SAR service
type Service struct {
	Name          string
	Version       string
	Actions       []Action
	ConditionKeys []Condition
	Resources     []Resource
}

// Action represents a SAR action
type Action struct {
	Name                string
	Service             string // technically doesn't exist, but we add this
	ActionConditionKeys []string
	// TODO(nsiow) implement resource checks in the simulator
	Resources []ResourcePointer
}

// ShortName provides the :-contatenated string representation of the action
func (a *Action) ShortName() string {
	if len(a.Service) == 0 || len(a.Name) == 0 {
		return ""
	}

	return a.Service + ":" + a.Name
}

// Condition represents a SAR condition
type Condition struct {
	Name  string
	Types []string
}

// Resource represents a SAR resource
type Resource struct {
	Name          string
	ARNFormats    []string
	ConditionKeys []string
}

// ResourcePointer represents a SAR resource pointer
type ResourcePointer struct {
	Name string
}
