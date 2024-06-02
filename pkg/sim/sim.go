package sim

import (
	"errors"
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
)

// Simulator provides the ability to simulate IAM policies and the interactions between
// Principals + Resources
type Simulator struct {
	options options
}

// NewSimulator creates and returns a Simulator with the provided options
func NewSimulator(o ...Option) (*Simulator, error) {
	s := Simulator{}

	// Execute any provided options
	var opts options
	for _, opt := range o {
		err := opt(&opts)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("error executing simulator options"), err)
		}
	}
	s.options = opts

	return &s, nil
}

// ComputeAccessSummary generates a numerical summary of access within the provided Environment
//
// The summary is returned in a map of format map[<resource_arn>]: <# of principals with access>
// where access is defined as any of the provided actions being allowed
func (s *Simulator) ComputeAccessSummary(
	env *entities.Environment,
	actions []string) (map[string]int, error) {
	// TODO(nsiow) this needs to be parallelized
	// Iterate over the matrix of Resources x Principals x Actions
	access := make(map[string]int)
	for _, r := range env.Resources {
		for _, p := range env.Principals {
			for _, a := range actions {
				result, err := s.Simulate(a, &p, &r, &AuthContext{})
				if err != nil {
					return nil, errors.Join(fmt.Errorf("error during simulation"), err)
				}

				if result.IsAllowed {
					access[r.Arn] += 1
					break
				}
			}
		}
	}

	return access, nil
}

// Simulate determines whether the provided Principal is able to perform the given Action on the
// specified Resource
func (s *Simulator) Simulate(
	action string,
	p *entities.Principal,
	r *entities.Resource,
	ac *AuthContext) (*Result, error) {
	nye
}
