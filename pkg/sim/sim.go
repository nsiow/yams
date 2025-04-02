package sim

import (
	"errors"
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
)

// Simulator provides the ability to simulate IAM policies and the interactions between
// Principals + Resources
type Simulator struct {
	universe entities.Universe
	options  Options
}

// NewSimulator creates and returns a Simulator with the provided options
func NewSimulator(o ...OptionF) (*Simulator, error) {
	s := Simulator{}

	// Execute any provided options
	var opts Options
	for _, opt := range o {
		err := opt(&opts)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("error executing simulator options"), err)
		}
	}
	s.options = opts

	return &s, nil
}

// Universe returns a pointer to the current Universe being used by the Simulator
func (s *Simulator) Universe() entities.Universe {
	return s.universe
}

// SetUniverse redefines the Universe used by the Simulator for access evaluations
func (s *Simulator) SetUniverse(universe entities.Universe) {
	s.universe = universe
}

// Simulate determines whether the provided AuthContext would be allowed
func (s *Simulator) Simulate(ac AuthContext) (*Result, error) {
	// TODO(nsiow) perform AuthContext validation
	return evalOverallAccess(&s.options, ac)
}

// SimulateByArn determines whether the operation would be allowed
func (s *Simulator) SimulateByArn(action, principal, resource string, ctx map[string]string) (*Result, error) {

	ac := AuthContext{}
	ac.Action = action
	ac.Properties = ctx

	// Locate Principal
	found := false
	for _, p := range s.universe.Principals {
		if p.Arn == principal {
			ac.Principal = p
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("simulator universe does not have Principal with Arn=%s", principal)
	}

	// Locate resource
	found = false
	for _, r := range s.universe.Resources {
		if r.Arn == resource {
			ac.Resource = r
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("simulator universe does not have Resource with Arn=%s", resource)
	}

	return s.Simulate(ac)
}

// ComputeAccessSummary generates a numerical summary of access within the provided Universe
//
// The summary is returned in a map of format map[<resource_arn>]: <# of principals with access>
// where access is defined as any of the provided actions being allowed
func (s *Simulator) ComputeAccessSummary(actions []string) (map[string]int, error) {
	// TODO(nsiow) this needs to be parallelized
	// Iterate over the matrix of Resources x Principals x Actions
	access := make(map[string]int)
	for _, r := range s.universe.Resources {
		// we do this because we always want resources to show up, regardless of access
		access[r.Arn] = 0

		for _, p := range s.universe.Principals {
			for _, a := range actions {
				ac := AuthContext{
					Action:     a,
					Principal:  p,
					Resource:   r,
					Properties: map[string]string{}}
				result, err := s.Simulate(ac)
				if err != nil {
					return nil, errors.Join(fmt.Errorf("error during simulation"), err)
				}

				if result.IsAllowed {
					fmt.Printf("access allowed between %s and %s\n", r.Arn, p.Arn)
					access[r.Arn]++
					break
				}
			}
		}
	}

	return access, nil
}
