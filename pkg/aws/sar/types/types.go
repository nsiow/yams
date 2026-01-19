package types

import (
	"strings"

	"github.com/nsiow/yams/pkg/sim/wildcard"
)

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
	AccessLevel         string
	ActionConditionKeys []string
	Resources           []Resource `json:"ResolvedResources"`
}

// ShortName provides the :-contatenated string representation of the action
func (a *Action) ShortName() string {
	return a.Service + ":" + a.Name
}

// HasTargets returns whether this Action targets any Resources
func (a *Action) HasTargets() bool {
	return len(a.Resources) > 0
}

// Targets determines whether this Action supports targeting of the specified Resource
func (a *Action) Targets(arn string) bool {
	for _, allowedResource := range a.Resources {
		// Check custom handling rules before format matching
		skip := false
		for _, handling := range allowedResource.CustomHandling {
			switch handling {
			case "DisallowSlashes":
				if strings.Contains(arn, "/") {
					skip = true
				}
			default:
				panic("unknown custom handling value: " + handling)
			}
		}
		if skip {
			continue
		}

		for _, allowedFormat := range allowedResource.ARNFormats {
			if wildcard.MatchSegments(allowedFormat, arn) {
				return true
			}
		}
	}

	return false
}

// Condition represents a SAR condition
type Condition struct {
	Name  string
	Types []string
}

// Resource represents a SAR resource
type Resource struct {
	Name           string
	ARNFormats     []string
	ConditionKeys  []string
	CustomHandling []string
}

// ResourcePointer represents a SAR resource pointer
type ResourcePointer struct {
	Name string
}

// TODO(nsiow) helper function for whether or not the action is one that applies to 0 resources,
// some resources, multi resources, etc
