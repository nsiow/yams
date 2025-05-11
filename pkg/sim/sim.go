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
	uv      *entities.Universe
	options *Options
}

// NewSimulator creates and returns a Simulator with the provided options
func NewSimulator(opts ...OptionF) (*Simulator, error) {
	s := Simulator{}
	s.options = NewOptions(opts...)

	return &s, nil
}

// Universe returns a pointer to the current Universe being used by the Simulator
func (s *Simulator) Universe() *entities.Universe {
	return s.uv
}

// SetUniverse redefines the Universe used by the Simulator for access evaluations
func (s *Simulator) SetUniverse(uv *entities.Universe) {
	s.uv = uv
}

// Simulate determines whether the provided AuthContext would be allowed
func (s *Simulator) Simulate(ac AuthContext) (*SimResult, error) {
	return s.SimulateWithOptions(ac, s.options)
}

// Simulate determines whether the provided AuthContext would be allowed
func (s *Simulator) SimulateWithOptions(ac AuthContext, opts *Options) (*SimResult, error) {
	err := ac.Validate()
	if err != nil {
		return nil, err
	}

	subj := newSubject(&ac, opts)
	return evalOverallAccess(subj)
}

// SimulateByArn determines whether the operation would be allowed between the Principal and
// Resource specified by the provided ARNs, using the Simulator's default options
func (s *Simulator) SimulateByArn(
	principalArn, action, resourceArn string,
	ctx map[string]string) (*SimResult, error) {
	return s.SimulateByArnWithOptions(principalArn, action, resourceArn, ctx, s.options)
}

// SimulateByArnWithOptions determines whether the operation would be allowed between the Principal
// and Resource specified by the provided ARNs, using the provided simulation Options
func (s *Simulator) SimulateByArnWithOptions(
	principalArn, action, resourceArn string,
	ctx map[string]string,
	opts *Options) (*SimResult, error) {

	var err error
	ac := AuthContext{}
	ac.Properties = NewBagFromMap(ctx)

	if resolvedAction, ok := sar.LookupString(action); !ok {
		return nil, fmt.Errorf("unable to resolve action '%s'", action)
	} else {
		ac.Action = resolvedAction
	}

	// Locate Principal
	p, ok := s.uv.Principal(principalArn)
	if !ok {
		return nil, fmt.Errorf("no principal with arn: %s", principalArn)
	}
	fp, err := p.Freeze()
	if err != nil {
		return nil, fmt.Errorf("error while freezing principal for simulation: %w", err)
	}
	ac.Principal = &fp

	// Locate Resource
	r, ok := s.uv.Resource(resourceArn)
	if !ok {
		return nil, fmt.Errorf("no resource with arn: %s", resourceArn)
	}
	fr, err := r.Freeze()
	if err != nil {
		return nil, fmt.Errorf("error while freezing resource for simulation: %w", err)
	}
	ac.Resource = &fr

	return s.Simulate(ac)
}

// ComputeAccessSummary generates a numerical summary of access within the provided Universe
//
// The summary is returned in a map of format map[<resource_arn>]: <# of principals with access>
// where access is defined as any of the provided actions being allowed
func (s *Simulator) ComputeAccessSummary(actions []*types.Action) (map[string]int, error) {
	ps, err := s.uv.FrozenPrincipals()
	if err != nil {
		return nil, fmt.Errorf("error while freezing principals for simulation: %w", err)
	}

	rs, err := s.uv.FrozenResources()
	if err != nil {
		return nil, fmt.Errorf("error while freezing resources for simulation: %w", err)
	}

	// TODO(nsiow) this needs to be parallelized
	// Iterate over the matrix of Resources x Principals x Actions
	access := make(map[string]int)
	for _, r := range rs {
		// we do this because we always want resources to show up, even if nothing can access it
		access[r.Arn] = 0

		for _, p := range ps {
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
					access[r.Arn]++
					break
				}
			}
		}
	}

	return access, nil
}
