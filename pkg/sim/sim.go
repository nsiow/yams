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
	options SimOptions
}

// NewSimulator creates and returns a Simulator with the provided options
func NewSimulator(o ...Option) (*Simulator, error) {
	s := Simulator{}

	// Execute any provided options
	var opts SimOptions
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

// SimulateEvent determines whether the provided Event would be allowed
func (s *Simulator) SimulateEvent(evt *Event) (*Result, error) {
	return evalOverallAccess(&s.options, evt)
}

// SimulateByArn determines whether the operation would be allowed
func (s *Simulator) SimulateByArn(action, principal, resource string, ac *AuthContext) (*Result, error) {
	evt := Event{}
	evt.Action = action
	evt.AuthContext = ac

	// Locate Principal
	for _, p := range s.env.Principals {
		if p.Arn == principal {
			evt.Principal = &p
			break
		}
	}
	if evt.Principal == nil {
		return nil, fmt.Errorf("simulator environment does not have Principal with Arn=%s", principal)
	}

	// Locate resource
	for _, r := range s.env.Resources {
		if r.Arn == resource {
			evt.Resource = &r
			break
		}
	}
	if evt.Resource == nil {
		return nil, fmt.Errorf("simulator environment does not have Resource with Arn=%s", resource)
	}

	return evalOverallAccess(&s.options, &evt)
}

// ComputeAccessSummary generates a numerical summary of access within the provided Environment
//
// The summary is returned in a map of format map[<resource_arn>]: <# of principals with access>
// where access is defined as any of the provided actions being allowed
func (s *Simulator) ComputeAccessSummary(actions []string) (map[string]int, error) {
	// TODO(nsiow) this needs to be parallelized
	// Iterate over the matrix of Resources x Principals x Actions
	access := make(map[string]int)
	for _, r := range s.env.Resources {
		for _, p := range s.env.Principals {
			for _, a := range actions {
				result, err := s.SimulateEvent(
					&Event{a, &p, &r, &AuthContext{}},
				)
				if err != nil {
					return nil, errors.Join(fmt.Errorf("error during simulation"), err)
				}

				if result.IsAllowed {
					access[r.Arn]++
					break
				}
			}
		}
	}

	return access, nil
}
