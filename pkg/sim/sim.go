package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/aws/sar/types"
	"github.com/nsiow/yams/pkg/entities"
)

// Simulator provides the ability to simulate IAM policies and the interactions between
// Principals + Resources
type Simulator struct {
	universe *entities.Universe
	options  Options
}

// NewSimulator creates and returns a Simulator with the provided options
func NewSimulator(o ...OptionF) (*Simulator, error) {
	s := Simulator{}
	s.options = *NewOptions(o...)

	return &s, nil
}

// Universe returns a pointer to the current Universe being used by the Simulator
func (s *Simulator) Universe() *entities.Universe {
	return s.universe
}

// SetUniverse redefines the Universe used by the Simulator for access evaluations
func (s *Simulator) SetUniverse(universe *entities.Universe) {
	s.universe = universe
}

// Simulate determines whether the provided AuthContext would be allowed
func (s *Simulator) Simulate(ac AuthContext) (*Result, error) {
	err := ac.Validate()
	if err != nil {
		return nil, err
	}

	subj := newSubject(&ac, &s.options)
	return evalOverallAccess(subj)
}

// SimulateByArn determines whether the operation would be allowed between the Principal and
// Resource specified by the provided ARN strings
func (s *Simulator) SimulateByArnString(
	action string,
	principalString string,
	resourceString string,
	ctx map[string]string) (*Result, error) {
	return s.SimulateByArn(action, entities.Arn(principalString), entities.Arn(resourceString), ctx)
}

// SimulateByArn determines whether the operation would be allowed between the Principal and
// Resource specified by the provided ARNs
func (s *Simulator) SimulateByArn(
	action string,
	principalArn entities.Arn,
	resourceArn entities.Arn,
	ctx map[string]string) (*Result, error) {

	ac := AuthContext{}
	ac.Properties = NewBagFromMap(ctx)

	if resolvedAction, ok := sar.LookupString(action); !ok {
		return nil, fmt.Errorf("unable to resolve action '%s'", action)
	} else {
		ac.Action = resolvedAction
	}

	// Locate Principal
	principal, ok := s.universe.Principal(principalArn)
	if !ok {
		return nil, fmt.Errorf("simulator universe does not have Principal with Arn=%s", principal)
	}
	ac.Principal = principal

	// Locate resource
	resource, ok := s.universe.Resource(resourceArn)
	if !ok {
		return nil, fmt.Errorf("simulator universe does not have Resource with Arn=%s", resource)
	}
	ac.Resource = resource

	return s.Simulate(ac)
}

// ComputeAccessSummary generates a numerical summary of access within the provided Universe
//
// The summary is returned in a map of format map[<resource_arn>]: <# of principals with access>
// where access is defined as any of the provided actions being allowed
func (s *Simulator) ComputeAccessSummary(actions []*types.Action) (map[string]int, error) {
	// TODO(nsiow) this needs to be parallelized
	// Iterate over the matrix of Resources x Principals x Actions
	access := make(map[string]int)
	for r := range s.universe.Resources() {
		// we do this because we always want resources to show up, even if nothing can access it
		access[r.Arn.String()] = 0

		for p := range s.universe.Principals() {
			for _, a := range actions {
				ac := AuthContext{
					Action:     a,
					Principal:  &p,
					Resource:   &r,
					Properties: NewBag[string]()}

				// Attempt to simulate, discard result on error
				result, err := s.Simulate(ac)
				if err != nil {
					continue
				}

				if result.IsAllowed {
					access[r.Arn.String()]++
					break
				}
			}
		}
	}

	return access, nil
}
