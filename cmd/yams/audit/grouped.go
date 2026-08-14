package audit

import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/sim"
)

type groupedWriter interface {
	WriteGroup(resource, group string, principals []string) error
	Flush() error
}

// processGroupedEntry simulates one resource type in bounded batches and emits complete groups.
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

	principalIndexes := make(map[string]int, len(frozenPrincipals))
	for i, principal := range frozenPrincipals {
		principalIndexes[principal.Arn] = i
	}

	actionGroups := make(map[string]int)
	for groupIndex, group := range entry.ActionGroups {
		for _, action := range group.Actions {
			resolved, ok := sar.LookupString(action)
			if !ok {
				return 0, 0, fmt.Errorf("unknown action %q", action)
			}
			actionGroups[resolved.ShortName()] = groupIndex
		}
	}

	slog.Info("processing grouped entry",
		"type", entry.ResourceType,
		"resources", len(resourceArns),
		"actions", len(entry.allActions()),
		"groups", len(entry.ActionGroups),
		"principals", len(frozenPrincipals),
		"batch_size", batchSize)

	var totalGroups, totalRelationships int
	for start := 0; start < len(resourceArns); start += batchSize {
		end := min(start+batchSize, len(resourceArns))
		groups, relationships, err := processGroupedBatch(
			simulator,
			frozenPrincipals,
			principalIndexes,
			resourceArns[start:end],
			entry,
			actionGroups,
			opts,
			w,
		)
		if err != nil {
			return 0, 0, err
		}
		totalGroups += groups
		totalRelationships += relationships
	}

	slog.Info("grouped entry complete",
		"type", entry.ResourceType,
		"groups", totalGroups,
		"relationships", totalRelationships)

	return totalGroups, totalRelationships, nil
}

func processGroupedBatch(
	simulator *sim.Simulator,
	frozenPrincipals []*entities.FrozenPrincipal,
	principalIndexes map[string]int,
	resourceArns []string,
	entry ConfigEntry,
	actionGroups map[string]int,
	opts sim.Options,
	w groupedWriter,
) (int, int, error) {
	expanded, err := simulator.ExpandResources(resourceArns, opts)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to expand resources: %w", err)
	}

	frozenResources, err := simulator.FreezeResources(expanded, opts)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to freeze resources: %w", err)
	}

	resourceIndexes := make(map[string]int, len(resourceArns))
	for i, resource := range resourceArns {
		resourceIndexes[resource] = i
	}

	wordCount := (len(frozenPrincipals) + 63) / 64
	membership := make([][][]uint64, len(resourceArns))
	for resourceIndex := range membership {
		membership[resourceIndex] = make([][]uint64, len(entry.ActionGroups))
		for groupIndex := range membership[resourceIndex] {
			membership[resourceIndex][groupIndex] = make([]uint64, wordCount)
		}
	}

	var resultErr error
	err = simulator.ProductFrozenStreaming(
		frozenPrincipals,
		entry.allActions(),
		frozenResources,
		opts,
		func(tuple sim.AccessTuple) {
			if resultErr != nil {
				return
			}

			groupIndex, ok := actionGroups[tuple.Action]
			if !ok {
				resultErr = fmt.Errorf("action %q is not mapped to a group", tuple.Action)
				return
			}
			resourceIndex, ok := resourceIndexes[collapseS3Arn(tuple.Resource)]
			if !ok {
				resultErr = fmt.Errorf("result resource %q is not in the current batch", tuple.Resource)
				return
			}
			principalIndex, ok := principalIndexes[tuple.Principal]
			if !ok {
				resultErr = fmt.Errorf("result principal %q is not frozen", tuple.Principal)
				return
			}

			membership[resourceIndex][groupIndex][principalIndex/64] |= 1 << (principalIndex % 64)
		},
	)
	if err != nil {
		return 0, 0, fmt.Errorf("simulation error: %w", err)
	}
	if resultErr != nil {
		return 0, 0, resultErr
	}

	var relationships int
	for resourceIndex, resource := range resourceArns {
		for groupIndex, group := range entry.ActionGroups {
			principals := principalsFromBits(membership[resourceIndex][groupIndex], frozenPrincipals)
			if err := w.WriteGroup(resource, group.Name, principals); err != nil {
				return 0, 0, fmt.Errorf("unable to write group: %w", err)
			}
			relationships += len(principals)
		}
	}

	return len(resourceArns) * len(entry.ActionGroups), relationships, nil
}

func principalsFromBits(bits []uint64, frozenPrincipals []*entities.FrozenPrincipal) []string {
	principals := make([]string, 0)
	for principalIndex, principal := range frozenPrincipals {
		if bits[principalIndex/64]&(1<<(principalIndex%64)) != 0 {
			principals = append(principals, principal.Arn)
		}
	}
	return principals
}
