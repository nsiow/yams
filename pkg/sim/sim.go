package sim

import (
	"errors"
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
)

// Simulator provides the ability to simulate IAM policies and the interactions between
// Principals + Resources
type Simulator struct {
	env     *entities.Environment
	options Options
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

// Environment returns a pointer to the current Environment being used by the Simulator
func (s *Simulator) Environment() *entities.Environment {
	return s.env
}

// SetEnvironment redefines the Environment used by the Simulator for access evaluations
func (s *Simulator) SetEnvironment(env *entities.Environment) {
	s.env = env
}

// Simulate determines whether the provided AuthContext would be allowed
func (s *Simulator) Simulate(ac AuthContext) (*Result, error) {
	// TODO(nsiow) perform AuthContext validation
	return evalOverallAccess(&s.options, ac)
}

// SimulateByArn determines whether the operation would be allowed
func (s *Simulator) SimulateByArn(action, principal, resource string, ctx map[string]string) (*Result, error) {

	// Validate that an Environment was set previously
	if s.env == nil {
		return nil, fmt.Errorf("Simulator has no environment set; use SetEnvironment(...) first")
	}

	ac := AuthContext{}
	ac.Action = action
	ac.Properties = ctx

	// Locate Principal
	for _, p := range s.env.Principals {
		if p.Arn == principal {
			ac.Principal = &p
			break
		}
	}
	if ac.Principal == nil {
		return nil, fmt.Errorf("simulator environment does not have Principal with Arn=%s", principal)
	}

	// Locate resource
	for _, r := range s.env.Resources {
		if r.Arn == resource {
			ac.Resource = &r
			break
		}
	}
	if ac.Resource == nil {
		return nil, fmt.Errorf("simulator environment does not have Resource with Arn=%s", resource)
	}

	return s.Simulate(ac)
}

// ComputeAccessSummary generates a numerical summary of access within the provided Environment
//
// The summary is returned in a map of format map[<resource_arn>]: <# of principals with access>
// where access is defined as any of the provided actions being allowed
func (s *Simulator) ComputeAccessSummary(actions []string) (map[string]int, error) {
	// Validate that an Environment was set previously
	if s.env == nil {
		return nil, fmt.Errorf("Simulator has no environment set; use SetEnvironment(...) first")
	}

	// TODO(nsiow) this needs to be parallelized
	// Iterate over the matrix of Resources x Principals x Actions
	access := make(map[string]int)
	for _, r := range s.env.Resources {
		// we do this because we always want resources to show up, regardless of access
		access[r.Arn] = 0

		for _, p := range s.env.Principals {
			for _, a := range actions {
				ac := AuthContext{
					Action:     a,
					Principal:  &p,
					Resource:   &r,
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
