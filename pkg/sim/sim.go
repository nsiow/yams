package sim

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nsiow/yams/internal/common"
	"github.com/nsiow/yams/pkg/arn"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/aws/sar/types"
	"github.com/nsiow/yams/pkg/entities"
)

// Simulator provides the ability to simulate IAM policies and the interactions between
// Principals + Resources
type Simulator struct {
	Universe *entities.Universe
	Pool     *Pool
}

// NewSimulator creates and returns a Simulator with the provided options
func NewSimulator() (*Simulator, error) {
	s := Simulator{}
	s.Universe = entities.NewUniverse()
	s.Universe.LoadBasePolicies()
	s.Pool = NewPool(context.TODO(), &s)
	s.Pool.Start()

	return &s, nil
}

// TODO(nsiow) move Universe/Options behind getters and setters

// resolvePrincipal finds and freezes a Principal through all overlays, indirections, etc
func (s *Simulator) resolvePrincipal(arn string, opts Options) (*entities.FrozenPrincipal, error) {
	uvs := s.Universe.Overlay(opts.Overlay)

	// first try exact match
	for _, uv := range uvs {
		principal, ok := uv.Principal(arn)
		if ok {
			fp, err := principal.FreezeWith(opts.Strict, uvs...)
			return &fp, err
		}
	}

	// then try fuzzy finding if enabled
	if opts.EnableFuzzyMatchArn {
		var matches []string
		for _, uv := range uvs {
			for _, principalArn := range uv.PrincipalArns() {
				if strings.Contains(strings.ToLower(principalArn), strings.ToLower(arn)) {
					if len(matches) < 10 {
						matches = append(matches, principalArn)
					}
				}
			}
		}

		if len(matches) == 1 {
			return s.resolvePrincipal(matches[0], opts)
		} else if len(matches) > 1 {
			return nil, fmt.Errorf("too many matches for '%s': %v", arn, matches)
		}
	}

	return nil, fmt.Errorf("no principal with arn: %s", arn)
}

// resolveResource finds and freezes a Resource through all overlays, indirections, etc
func (s *Simulator) resolveResource(arn string, opts Options) (*entities.FrozenResource, error) {
	uvs := s.Universe.Overlay(opts.Overlay)

	// first try exact match
	for _, uv := range uvs {
		resource, ok := uv.Resource(arn)
		if ok {
			fr, err := resource.FreezeWith(opts.Strict, uvs...)
			return &fr, err
		}
	}

	// then try fuzzy finding if enabled
	if opts.EnableFuzzyMatchArn {
		var matches []string
		for _, uv := range uvs {
			for _, resourceArn := range uv.ResourceArns() {
				if strings.Contains(strings.ToLower(resourceArn), strings.ToLower(arn)) {
					if len(matches) < 10 {
						matches = append(matches, resourceArn)
					}
				}
			}
		}

		if len(matches) == 1 {
			return s.resolveResource(matches[0], opts)
		} else if len(matches) > 1 {
			return nil, fmt.Errorf("too many matches for '%s': %v", arn, matches)
		}
	}

	return nil, fmt.Errorf("no resource with arn: %s", arn)
}

// expandResources takes the provided list of Resource ARNs and specified options, and performs any
// required expansion of Resources into Sub-resources. For example, expanding a resource set with
// a non-empty value for DefaultS3Key will add a new Resource to the set for each S3 bucket.
//
// TODO(nsiow) revisit this implementation
func (s *Simulator) expandResources(arns []string, opts Options) ([]string, error) {
	expanded := make([]string, 0)

	for _, arn := range arns {
		expanded = append(expanded, arn)

		if opts.DefaultS3Key != "" &&
			strings.HasPrefix(arn, "arn:aws:s3:::") &&
			!strings.Contains(arn, "/") {
			resource, ok := s.Universe.Resource(arn)
			if !ok {
				return nil, fmt.Errorf("unable to locate resource for expansion: '%s'", arn)
			}

			subresource, err := resource.SubResource(opts.DefaultS3Key)
			if err != nil {
				return nil, err
			}

			expanded = append(expanded, subresource.Arn)
		}
	}

	return expanded, nil
}

// Simulate determines whether the provided AuthContext would be allowed
func (s *Simulator) Simulate(ac AuthContext) (*SimResult, error) {
	return s.SimulateWithOptions(ac, DEFAULT_OPTIONS)
}

// Simulate determines whether the provided AuthContext would be allowed
func (s *Simulator) SimulateWithOptions(ac AuthContext, opts Options) (*SimResult, error) {
	if opts.ForceFailure {
		return nil, fmt.Errorf("error due to forced-failure option")
	}

	err := ac.Validate(opts)
	if err != nil {
		return nil, err
	}

	// TODO(nsiow) see if we can add stronger guarantees around P/A/R being set
	subj := newSubject(ac, opts)
	result := evalOverallAccess(subj)
	result.Principal = ac.Principal.Arn
	result.Action = ac.Action.ShortName()
	if ac.Resource != nil {
		result.Resource = ac.Resource.Arn
	}

	return result, nil
}

// SimulateByArn determines whether the operation would be allowed between the Principal and
// Resource specified by the provided ARNs, using the Simulator's default options
func (s *Simulator) SimulateByArn(principalArn, action, resourceArn string) (*SimResult, error) {
	return s.SimulateByArnWithOptions(principalArn, action, resourceArn, DEFAULT_OPTIONS)
}

// SimulateByArnWithOptions determines whether the operation would be allowed between the Principal
// and Resource specified by the provided ARNs, using the provided simulation Options
func (s *Simulator) SimulateByArnWithOptions(
	principalArn, action, resourceArn string, opts Options) (*SimResult, error) {

	var err error
	ac := AuthContext{}
	ac.Properties = opts.Context

	if resolvedAction, ok := sar.LookupString(action); !ok {
		return nil, fmt.Errorf("unable to resolve action '%s'", action)
	} else {
		ac.Action = resolvedAction
	}

	// Locate Principal
	fp, err := s.resolvePrincipal(principalArn, opts)
	if err != nil {
		return nil, fmt.Errorf("error resolving principal for simulation: %w", err)
	}
	ac.Principal = fp

	// Locate Resource (if needed)
	if ac.Action.HasTargets() {
		_, ok := s.Universe.Resource(resourceArn)
		if !ok && strings.HasPrefix(ac.Action.Name, "Create") {
			// Handle case where API call DOES have targets but those targets shouldn't exist yet. It's
			// really just targeted Create* calls
			// TODO(nsiow) revisit this
			ac.Resource = &entities.FrozenResource{
				AccountId: fp.AccountId,
				Arn:       resourceArn,
			}
		} else {
			// Handle normal case where API call does have targets and also those targets should exist
			fr, err := s.resolveResource(resourceArn, opts)
			if err != nil {
				return nil, fmt.Errorf("error resolving resource for simulation: %w", err)
			}
			ac.Resource = fr
		}
	}

	return s.SimulateWithOptions(ac, opts)
}

func (s *Simulator) WhichPrincipals(action, resource string, opts Options) ([]string, error) {
	matrix, err := s.Product(
		s.Universe.PrincipalArns(),
		[]string{action},
		[]string{resource},
		opts,
	)
	if err != nil {
		return nil, err
	}

	allowed := []string{}
	for _, tuple := range matrix {
		if tuple.Result.IsAllowed {
			allowed = append(allowed, tuple.Principal)
		}
	}
	return allowed, nil
}

func (s *Simulator) WhichActions(principal, resource string, opts Options) ([]string, error) {
	svc := arn.Service(resource)
	actions := sar.NewQuery().WithService(svc).Results()

	matrix, err := s.Product(
		[]string{principal},
		common.Map(actions, func(a types.Action) string { return a.ShortName() }),
		[]string{resource},
		opts,
	)
	if err != nil {
		return nil, err
	}

	allowed := []string{}
	for _, tuple := range matrix {
		if tuple.Result.IsAllowed {
			allowed = append(allowed, tuple.Action)
		}
	}
	return allowed, nil
}

func (s *Simulator) WhichResources(principal, action string, opts Options) ([]string, error) {
	expandedResources := s.Universe.ResourceArns()
	expandedResources, err := s.expandResources(expandedResources, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to expand provided resource list: %w", err)
	}

	matrix, err := s.Product(
		[]string{principal},
		[]string{action},
		expandedResources,
		opts,
	)
	if err != nil {
		return nil, err
	}

	allowed := []string{}
	for _, tuple := range matrix {
		if tuple.Result.IsAllowed {
			allowed = append(allowed, tuple.Resource)
		}
	}
	return allowed, nil
}

func (s *Simulator) AccessSummary(actions []string, opts Options) (map[string]int, error) {
	resourceArns := s.Universe.ResourceArns()
	resourceArns, err := s.expandResources(resourceArns, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to expand provided resource list: %w", err)
	}

	matrix, err := s.Product(
		s.Universe.PrincipalArns(),
		actions,
		resourceArns,
		opts)
	if err != nil {
		return nil, err
	}

	access := make(map[string]map[string]bool)
	for _, tuple := range matrix {
		if _, ok := access[tuple.Resource]; !ok {
			access[tuple.Resource] = make(map[string]bool)
		}

		if tuple.Result.IsAllowed {
			access[tuple.Resource][tuple.Principal] = true
		}
	}

	summary := make(map[string]int)
	for _, arn := range s.Universe.ResourceArns() {
		summary[arn] = 0
	}
	for resource, principals := range access {
		summary[resource] = len(principals)
	}
	return summary, nil
}

type AccessTuple struct {
	Principal string
	Action    string
	Resource  string
	Result    *SimResult
}

// Product is a mostly-helper function (that can be used directly!) which calculates the Cartesian
// product of the provided simulation identifiers, while also filtering out any combinations that
// are not allowed.
func (s *Simulator) Product(ps, as, rs []string, opts Options) ([]AccessTuple, error) {
	var fas []*types.Action
	for _, a := range as {
		fa, ok := sar.LookupString(a)
		if !ok {
			return nil, fmt.Errorf("unknown action: %s", a)
		}
		fas = append(fas, fa)
	}

	var fps []*entities.FrozenPrincipal
	for _, p := range ps {
		fp, err := s.resolvePrincipal(p, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve principal '%s' for simulation: %w", p, err)
		}
		fps = append(fps, fp)
	}

	var frs []*entities.FrozenResource
	for _, r := range rs {
		fr, err := s.resolveResource(r, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve resource '%s' for simulation: %w", r, err)
		}
		frs = append(frs, fr)
	}

	var submitted int32
	var done atomic.Int32
	finished := make(chan simOut, s.Pool.NumWorkers())
	batch := simBatch{Jobs: []simIn{}, Finished: finished, Done: &done}

	for _, p := range fps {
		for _, a := range fas {
			for _, r := range frs {
				// Skip invalid combinations
				if !a.Targets(r.Arn) {
					continue
				}

				job := simIn{
					AuthContext{
						Action:     a,
						Principal:  p,
						Resource:   r,
						Properties: opts.Context,
					},
					opts,
				}
				batch.Jobs = append(batch.Jobs, job)
				submitted++

				if len(batch.Jobs) == s.Pool.BatchSize() {
					s.Pool.Submit(batch)
					batch = simBatch{Jobs: []simIn{}, Finished: finished}
				}
			}
		}
	}

	// leftovers
	if len(batch.Jobs) > 0 {
		s.Pool.Submit(batch)
	}

	var matrix []AccessTuple

	for done.Load() < submitted {
		select {
		case job := <-finished:
			if job.Error != nil {
				return nil, fmt.Errorf("simulation error: %w", job.Error)
			}

			if job.Result.IsAllowed { // should always be the case, but confirm
				matrix = append(matrix, AccessTuple{
					Principal: job.Result.Principal,
					Action:    job.Result.Action,
					Resource:  job.Result.Resource,
					Result:    job.Result})
			}
		case <-time.After(time.Second * 1):
			break
		}
	}

	close(finished)
	return matrix, nil
}
