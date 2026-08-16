package sim

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nsiow/yams/internal/common"
	"github.com/nsiow/yams/pkg/arn"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/aws/sar/types"
	"github.com/nsiow/yams/pkg/entities"
)

// placeholderWildcardAccount marks a placeholder resource whose target account is not
// pinned. At simulation time the account segment is filled in with the principal's account
// so each principal tests against a resource in their own account. A pinned account (e.g.
// arn:aws:iam::88888:role/x) keeps that value and enforces same-account via the normal
// evalIsSameAccount path.
const placeholderWildcardAccount = "*"

// isCreateAction returns true for actions that target resources which don't exist yet
func isCreateAction(a *types.Action) bool {
	return strings.HasPrefix(a.Name, "Create") || a.Name == "RunInstances"
}

// newPlaceholderResource builds a minimal FrozenResource for Create*/RunInstances targets.
// If the ARN has an explicit account segment, it's used as-is. Missing or wildcard account
// segments defer to per-principal substitution via specializeForPrincipal.
func newPlaceholderResource(resArn string) *entities.FrozenResource {
	acct := arn.Account(resArn)
	if acct == "" {
		acct = placeholderWildcardAccount
	}
	return &entities.FrozenResource{
		AccountId:   acct,
		Arn:         resArn,
		ArnSegments: entities.SplitArn(resArn),
	}
}

// specializeForPrincipal returns a copy of the resource with wildcard-account placeholders
// rewritten to use the principal's account. For resources whose account is already pinned,
// or non-placeholder resources, the input is returned unchanged.
func specializeForPrincipal(r *entities.FrozenResource, p *entities.FrozenPrincipal) *entities.FrozenResource {
	if r == nil || p == nil || r.AccountId != placeholderWildcardAccount {
		return r
	}
	newArn := rewriteAccountSegment(r.Arn, p.AccountId)
	return &entities.FrozenResource{
		AccountId:   p.AccountId,
		Arn:         newArn,
		ArnSegments: entities.SplitArn(newArn),
	}
}

// rewriteAccountSegment returns arnStr with the account segment (5th component) replaced by
// acct. ARNs with fewer than 5 segments are returned unchanged.
func rewriteAccountSegment(arnStr, acct string) string {
	parts := strings.SplitN(arnStr, ":", 6)
	if len(parts) < 5 {
		return arnStr
	}
	parts[4] = acct
	return strings.Join(parts, ":")
}

// allActionsAreCreate returns true if every action in the slice is a create-type action
func allActionsAreCreate(fas []*types.Action) bool {
	for _, a := range fas {
		if !isCreateAction(a) {
			return false
		}
	}
	return len(fas) > 0
}

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
	s.Pool = NewPool(context.Background(), &s)

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

// ExpandResources takes the provided list of Resource ARNs and performs any required expansion of
// Resources into Sub-resources (e.g. S3 bucket → object)
func (s *Simulator) ExpandResources(arns []string, opts Options) ([]string, error) {
	return s.expandResources(arns, opts)
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

// SimulateWithOptions determines whether the provided AuthContext would be allowed
func (s *Simulator) SimulateWithOptions(ac AuthContext, opts Options) (*SimResult, error) {
	if opts.ForceFailure {
		return nil, fmt.Errorf("error due to forced-failure option")
	}

	err := ac.Validate(opts)
	if err != nil {
		return nil, err
	}

	subj := newSubject(ac, opts)
	result := evalOverallAccess(&subj)
	result.Principal = ac.Principal.Arn
	result.Action = ac.Action.ShortName()
	if ac.Resource != nil {
		result.Resource = ac.Resource.Arn
	}
	if opts.EnableTracing {
		result.Trace = &subj.trc
	}

	return &result, nil
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
	ac.MultiValueProperties = opts.MultiValueContext

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
		if !ok && isCreateAction(ac.Action) {
			ac.Resource = specializeForPrincipal(newPlaceholderResource(resourceArn), ac.Principal)
		} else {
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
	actions := actionsForResource(resource)

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

func actionsForResource(resource string) []types.Action {
	svc := arn.Service(resource)
	seen := map[string]bool{}
	actions := []types.Action{}

	add := func(action types.Action) {
		key := action.ShortName()
		if seen[key] {
			return
		}
		seen[key] = true
		actions = append(actions, action)
	}

	for _, action := range sar.ActionsByService(svc) {
		add(action)
	}
	for _, action := range sar.AllActions() {
		if action.Targets(resource) {
			add(action)
		}
	}

	return actions
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

// IndexedAccess identifies an allowed product result by its input slice indexes.
type IndexedAccess struct {
	PrincipalIndex int
	ActionIndex    int
	ResourceIndex  int
}

func resolveActions(actions []string) ([]*types.Action, error) {
	resolved := make([]*types.Action, 0, len(actions))
	for _, action := range actions {
		value, ok := sar.LookupString(action)
		if !ok {
			return nil, fmt.Errorf("unknown action %q", action)
		}
		resolved = append(resolved, value)
	}
	return resolved, nil
}

// Product returns the allowed combinations in the Cartesian product of the supplied identifiers.
func (s *Simulator) Product(ps, as, rs []string, opts Options) ([]AccessTuple, error) {
	simId := rand.Text()
	slog.Debug("calculating product",
		"sim_id", simId)

	fas, err := resolveActions(as)
	if err != nil {
		return nil, err
	}

	fps, err := s.FreezePrincipals(ps, opts)
	if err != nil {
		return nil, err
	}

	frs, err := s.freezeResources(rs, opts, allActionsAreCreate(fas))
	if err != nil {
		return nil, err
	}

	slog.Debug("froze entities",
		"sim_id", simId)

	return s.collectProduct(fps, fas, frs, opts)
}

// FreezePrincipals resolves and freezes all the provided principal ARNs. This allows callers to
// freeze once and reuse across multiple Product calls.
func (s *Simulator) FreezePrincipals(arns []string, opts Options) ([]*entities.FrozenPrincipal, error) {
	fps := make([]*entities.FrozenPrincipal, 0, len(arns))
	for _, p := range arns {
		fp, err := s.resolvePrincipal(p, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve principal '%s': %w", p, err)
		}
		fps = append(fps, fp)
	}
	return fps, nil
}

// FreezeResources resolves and freezes all the provided resource ARNs
func (s *Simulator) FreezeResources(arns []string, opts Options) ([]*entities.FrozenResource, error) {
	return s.freezeResources(arns, opts, false)
}

// freezeResources resolves and freezes resource ARNs. When allowPlaceholders is true (i.e. all
// actions are Create*/RunInstances), unresolvable ARNs get a minimal placeholder instead of
// returning an error.
func (s *Simulator) freezeResources(arns []string, opts Options, allowPlaceholders bool) ([]*entities.FrozenResource, error) {
	frs := make([]*entities.FrozenResource, 0, len(arns))
	for _, r := range arns {
		fr, err := s.resolveResource(r, opts)
		if err != nil {
			if !allowPlaceholders {
				return nil, fmt.Errorf("unable to resolve resource '%s': %w", r, err)
			}
			fr = newPlaceholderResource(r)
		}
		frs = append(frs, fr)
	}
	return frs, nil
}

// ProductFrozenStreaming runs a Cartesian product simulation with pre-frozen entities and passes
// allowed results to onResult instead of collecting them in memory.
func (s *Simulator) ProductFrozenStreaming(
	fps []*entities.FrozenPrincipal,
	actions []string,
	frs []*entities.FrozenResource,
	opts Options,
	onResult func(AccessTuple),
) error {
	fas, err := resolveActions(actions)
	if err != nil {
		return err
	}

	return s.streamProduct(fps, fas, frs, opts, onResult)
}

// ProductFrozenIndexed runs a Cartesian product simulation and reports allowed combinations by
// their indexes in the supplied slices. The callback may run concurrently. Principal indexes that
// share a 64-index block are processed by the same goroutine, so callbacks can update a bitset
// indexed by PrincipalIndex without synchronization. The first callback error stops the simulation.
func (s *Simulator) ProductFrozenIndexed(
	ctx context.Context,
	fps []*entities.FrozenPrincipal,
	actions []string,
	frs []*entities.FrozenResource,
	opts Options,
	onResult func(IndexedAccess) error,
) error {
	fas, err := resolveActions(actions)
	if err != nil {
		return err
	}

	filtered := precomputeTargets(fas, frs)
	chunkSize := indexedPrincipalChunkSize(len(fps), s.Pool.NumWorkers())
	callbacks := productCallbacks{}
	if onResult != nil {
		callbacks.onAllowed = func(_ int, access IndexedAccess) error {
			return onResult(access)
		}
	}
	return s.runProduct(ctx, fps, filtered, opts, chunkSize, callbacks)
}

// collectProduct adapts the streaming product to the original collection API.
func (s *Simulator) collectProduct(
	fps []*entities.FrozenPrincipal,
	fas []*types.Action,
	frs []*entities.FrozenResource,
	opts Options,
) ([]AccessTuple, error) {
	var matrix []AccessTuple
	err := s.streamProduct(fps, fas, frs, opts, func(t AccessTuple) {
		matrix = append(matrix, t)
	})

	return matrix, err
}

type targetedResource struct {
	index    int
	resource *entities.FrozenResource
}

// actionResources pairs an action with the pre-filtered resources it can target.
type actionResources struct {
	index     int
	action    *types.Action
	resources []targetedResource
}

// precomputeTargets builds a filtered mapping of actions to their targeted resources. This moves
// target matching from O(principals x actions x resources) to O(actions x resources).
func precomputeTargets(fas []*types.Action, frs []*entities.FrozenResource) []actionResources {
	result := make([]actionResources, 0, len(fas))
	for actionIndex, a := range fas {
		var matching []targetedResource
		for resourceIndex, r := range frs {
			if a.Targets(r.Arn) {
				matching = append(matching, targetedResource{
					index:    resourceIndex,
					resource: r,
				})
			}
		}
		if len(matching) > 0 {
			result = append(result, actionResources{
				index:     actionIndex,
				action:    a,
				resources: matching,
			})
		}
	}
	return result
}

type accessIndexes struct {
	principal int
	action    int
	resource  int
}

type productCallbacks struct {
	onAllowed func(workerIndex int, access IndexedAccess) error
	onDone    func(workerIndex int)
}

const streamedResultBatchSize = 256

// streamProduct adapts the concurrent indexed engine to the original serialized callback API.
func (s *Simulator) streamProduct(
	fps []*entities.FrozenPrincipal,
	fas []*types.Action,
	frs []*entities.FrozenResource,
	opts Options,
	onResult func(AccessTuple),
) error {
	filtered := precomputeTargets(fas, frs)
	chunkSize := s.principalChunkSize(filtered)
	resultBatches := make(chan []accessIndexes, s.Pool.NumWorkers())
	simulationDone := make(chan error, 1)
	workerResults := make([][]accessIndexes, s.Pool.NumWorkers())

	go func() {
		defer close(resultBatches)
		simulationDone <- s.runProduct(
			context.Background(),
			fps,
			filtered,
			opts,
			chunkSize,
			productCallbacks{
				onAllowed: func(workerIndex int, access IndexedAccess) error {
					batch := append(workerResults[workerIndex], accessIndexes{
						principal: access.PrincipalIndex,
						action:    access.ActionIndex,
						resource:  access.ResourceIndex,
					})
					if len(batch) < streamedResultBatchSize {
						workerResults[workerIndex] = batch
						return nil
					}

					resultBatches <- batch
					workerResults[workerIndex] = nil
					return nil
				},
				onDone: func(workerIndex int) {
					if len(workerResults[workerIndex]) == 0 {
						return
					}
					resultBatches <- workerResults[workerIndex]
				},
			},
		)
	}()

	for batch := range resultBatches {
		results := make([]SimResult, len(batch))
		for i, access := range batch {
			principal := fps[access.principal]
			action := fas[access.action]
			resource := specializeForPrincipal(frs[access.resource], principal)
			result := &results[i]
			*result = SimResult{
				Principal: principal.Arn,
				Action:    action.ShortName(),
				Resource:  resource.Arn,
				IsAllowed: true,
			}
			onResult(AccessTuple{
				Principal: result.Principal,
				Action:    result.Action,
				Resource:  result.Resource,
				Result:    result,
			})
		}
	}

	return <-simulationDone
}

func (s *Simulator) principalChunkSize(filtered []actionResources) int {
	targetsPerPrincipal := 0
	for _, action := range filtered {
		targetsPerPrincipal += len(action.resources)
	}
	if targetsPerPrincipal == 0 {
		return 1
	}
	return max(1, s.Pool.BatchSize()/targetsPerPrincipal)
}

// indexedPrincipalChunkSize keeps large aggregations on cache-line-sized principal word ranges.
// Smaller products use narrower word-aligned ranges so they can still occupy all workers.
func indexedPrincipalChunkSize(principalCount, workerCount int) int {
	const (
		principalsPerWord = 64
		cacheLineWords    = 8
	)

	wordCount := (principalCount + principalsPerWord - 1) / principalsPerWord
	if wordCount == 0 || workerCount < 1 {
		return principalsPerWord
	}
	wordsPerChunk := min(cacheLineWords, max(1, wordCount/workerCount))
	return wordsPerChunk * principalsPerWord
}

// runProduct is the Cartesian product engine. Workers dynamically claim contiguous principal
// ranges, and each principal is evaluated by only one worker.
func (s *Simulator) runProduct(
	ctx context.Context,
	fps []*entities.FrozenPrincipal,
	filtered []actionResources,
	opts Options,
	chunkSize int,
	callbacks productCallbacks,
) error {
	if opts.ForceFailure {
		return fmt.Errorf("error due to forced-failure option")
	}
	if len(fps) == 0 || len(filtered) == 0 {
		return nil
	}
	runCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	if s.Pool.Ctx != nil {
		if err := context.Cause(s.Pool.Ctx); err != nil {
			cancel(err)
		}
		stopPoolCancel := context.AfterFunc(s.Pool.Ctx, func() {
			cancel(context.Cause(s.Pool.Ctx))
		})
		defer stopPoolCancel()
	}
	if err := context.Cause(runCtx); err != nil {
		return err
	}

	chunkSize = max(1, chunkSize)
	chunkCount := (len(fps) + chunkSize - 1) / chunkSize
	workerCount := min(s.Pool.NumWorkers(), chunkCount)
	logProgress := slog.Default().Enabled(runCtx, slog.LevelDebug)

	var nextPrincipal atomic.Uint64
	var processedPrincipals atomic.Uint64

	var workers sync.WaitGroup
	workers.Add(workerCount)
	for workerIndex := range workerCount {
		go func() {
			defer workers.Done()
			defer func() {
				if callbacks.onDone != nil {
					callbacks.onDone(workerIndex)
				}
			}()

			subj := newSubject(AuthContext{
				Properties:           opts.Context,
				MultiValueProperties: opts.MultiValueContext,
			}, opts)

			for {
				start := int(nextPrincipal.Add(uint64(chunkSize)) - uint64(chunkSize))
				if start >= len(fps) {
					return
				}
				end := min(start+chunkSize, len(fps))

				for principalIndex := start; principalIndex < end; principalIndex++ {
					if runCtx.Err() != nil {
						return
					}
					principal := fps[principalIndex]
					subj.auth.Principal = principal

					for _, action := range filtered {
						subj.auth.Action = action.action
						for _, target := range action.resources {
							if opts.EnableTracing {
								subj = newSubject(subj.auth, opts)
							}
							subj.auth.Resource = specializeForPrincipal(target.resource, principal)
							subj.extra = Extra{}
							subj.policyVersion = ""

							if !evalOverallAccess(&subj).IsAllowed || callbacks.onAllowed == nil {
								continue
							}
							if err := callbacks.onAllowed(workerIndex, IndexedAccess{
								PrincipalIndex: principalIndex,
								ActionIndex:    action.index,
								ResourceIndex:  target.index,
							}); err != nil {
								cancel(err)
								return
							}
						}
					}
				}
				if logProgress {
					processedPrincipals.Add(uint64(end - start))
				}
			}
		}()
	}

	if !logProgress {
		workers.Wait()
		return context.Cause(runCtx)
	}

	workersDone := make(chan struct{})
	go func() {
		workers.Wait()
		close(workersDone)
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-workersDone:
			return context.Cause(runCtx)
		case <-ticker.C:
			slog.Debug("simulation in progress",
				"processed_principals", processedPrincipals.Load(),
				"total_principals", len(fps))
		case <-runCtx.Done():
			<-workersDone
			return context.Cause(runCtx)
		}
	}
}
