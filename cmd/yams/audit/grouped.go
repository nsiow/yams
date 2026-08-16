package audit

import (
	"context"
	"fmt"
	"log/slog"
	"math/bits"
	"sort"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/sim"
)

type groupedWriter interface {
	WriteGroup(resource, group string, principals []string) error
	Flush() error
}

const groupedBatchQueueSize = 1

type groupedBatch struct {
	resources  []string
	membership []uint64
	wordCount  int
}

type groupedWriteResult struct {
	groups        int
	relationships int
	err           error
}

type groupedPlan struct {
	actions        []string
	groups         []ActionGroup
	groupForAction []int
}

func newGroupedPlan(groups []ActionGroup) groupedPlan {
	plan := groupedPlan{groups: groups}
	for groupIndex, group := range groups {
		for _, action := range group.Actions {
			plan.actions = append(plan.actions, action)
			plan.groupForAction = append(plan.groupForAction, groupIndex)
		}
	}
	return plan
}

// processGroupedEntry simulates one resource type in bounded batches. A single writer consumes
// completed batches in order while the next batch is simulated.
func processGroupedEntry(
	simulator *sim.Simulator,
	frozenPrincipals []*entities.FrozenPrincipal,
	entry ConfigEntry,
	opts sim.Options,
	batchSize int,
	w groupedWriter,
) (int, int, error) {
	var resourceArns []string
	for resource := range simulator.Universe.Resources() {
		if resource.Type == entry.ResourceType {
			resourceArns = append(resourceArns, resource.Arn)
		}
	}
	sort.Strings(resourceArns)

	if len(resourceArns) == 0 {
		slog.Info("no resources found for type", "type", entry.ResourceType)
		return 0, 0, nil
	}

	plan := newGroupedPlan(entry.ActionGroups)

	slog.Info("processing grouped entry",
		"type", entry.ResourceType,
		"resources", len(resourceArns),
		"actions", len(plan.actions),
		"groups", len(plan.groups),
		"principals", len(frozenPrincipals),
		"batch_size", batchSize)

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	batches := make(chan groupedBatch, groupedBatchQueueSize)
	writeDone := make(chan groupedWriteResult, 1)
	go func() {
		result := writeGroupedBatches(ctx, batches, frozenPrincipals, plan.groups, w)
		if result.err != nil {
			cancel(result.err)
		}
		writeDone <- result
	}()

	for start := 0; start < len(resourceArns); start += batchSize {
		end := min(start+batchSize, len(resourceArns))
		batch, err := processGroupedBatch(
			ctx,
			simulator,
			frozenPrincipals,
			resourceArns[start:end],
			plan,
			opts,
		)
		if err != nil {
			cancel(err)
			break
		}

		select {
		case batches <- batch:
		case <-ctx.Done():
		}
		if ctx.Err() != nil {
			break
		}
	}
	close(batches)
	writeResult := <-writeDone

	if writeResult.err != nil {
		return 0, 0, writeResult.err
	}
	if err := context.Cause(ctx); err != nil {
		return 0, 0, err
	}

	slog.Info("grouped entry complete",
		"type", entry.ResourceType,
		"groups", writeResult.groups,
		"relationships", writeResult.relationships)

	return writeResult.groups, writeResult.relationships, nil
}

func processGroupedBatch(
	ctx context.Context,
	simulator *sim.Simulator,
	frozenPrincipals []*entities.FrozenPrincipal,
	resourceArns []string,
	plan groupedPlan,
	opts sim.Options,
) (groupedBatch, error) {
	expanded, err := simulator.ExpandResources(resourceArns, opts)
	if err != nil {
		return groupedBatch{}, fmt.Errorf("unable to expand resources: %w", err)
	}

	frozenResources, err := simulator.FreezeResources(expanded, opts)
	if err != nil {
		return groupedBatch{}, fmt.Errorf("unable to freeze resources: %w", err)
	}

	resourceIndexByARN := make(map[string]int, len(resourceArns))
	for i, resource := range resourceArns {
		resourceIndexByARN[resource] = i
	}

	batchResourceIndexes := make([]int, len(frozenResources))
	for resourceIndex, resource := range frozenResources {
		batchResourceIndex, ok := resourceIndexByARN[collapseS3Arn(resource.Arn)]
		if !ok {
			return groupedBatch{}, fmt.Errorf(
				"expanded resource %q is not in the current batch",
				resource.Arn,
			)
		}
		batchResourceIndexes[resourceIndex] = batchResourceIndex
	}

	wordCount := (len(frozenPrincipals) + 63) / 64
	membership := make([]uint64, len(resourceArns)*len(plan.groups)*wordCount)

	err = simulator.ProductFrozenIndexed(
		ctx,
		frozenPrincipals,
		plan.actions,
		frozenResources,
		opts,
		func(access sim.IndexedAccess) error {
			groupIndex := plan.groupForAction[access.ActionIndex]
			resourceIndex := batchResourceIndexes[access.ResourceIndex]
			offset := (resourceIndex*len(plan.groups)+groupIndex)*wordCount +
				access.PrincipalIndex/64
			membership[offset] |= 1 << (access.PrincipalIndex % 64)
			return nil
		},
	)
	if err != nil {
		return groupedBatch{}, fmt.Errorf("simulation error: %w", err)
	}

	return groupedBatch{
		resources:  resourceArns,
		membership: membership,
		wordCount:  wordCount,
	}, nil
}

func writeGroupedBatches(
	ctx context.Context,
	batches <-chan groupedBatch,
	frozenPrincipals []*entities.FrozenPrincipal,
	groups []ActionGroup,
	w groupedWriter,
) groupedWriteResult {
	var result groupedWriteResult
	for {
		select {
		case <-ctx.Done():
			return result
		case batch, ok := <-batches:
			if !ok {
				return result
			}

			for resourceIndex, resource := range batch.resources {
				for groupIndex, group := range groups {
					start := (resourceIndex*len(groups) + groupIndex) * batch.wordCount
					end := start + batch.wordCount
					principals := principalsFromBits(
						batch.membership[start:end],
						frozenPrincipals,
					)
					if err := w.WriteGroup(resource, group.Name, principals); err != nil {
						result.err = fmt.Errorf("unable to write group: %w", err)
						return result
					}
					result.groups++
					result.relationships += len(principals)
				}
			}
		}
	}
}

func principalsFromBits(
	membership []uint64,
	frozenPrincipals []*entities.FrozenPrincipal,
) []string {
	principalCount := 0
	for _, word := range membership {
		principalCount += bits.OnesCount64(word)
	}

	principals := make([]string, 0, principalCount)
	for wordIndex, word := range membership {
		for word != 0 {
			bitIndex := bits.TrailingZeros64(word)
			principalIndex := wordIndex*64 + bitIndex
			if principalIndex < len(frozenPrincipals) {
				principals = append(principals, frozenPrincipals[principalIndex].Arn)
			}
			word &= word - 1
		}
	}
	return principals
}
