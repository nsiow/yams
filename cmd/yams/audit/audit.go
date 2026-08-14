package audit

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"
	"runtime/pprof"
	"sort"
	"strings"

	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/internal/smartrw"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/server"
	"github.com/nsiow/yams/pkg/sim"
)

const (
	formatCSV          = "csv"
	formatGroupedCSV   = "grouped-csv"
	formatGroupedJSONL = "grouped-jsonl"
)

// ActionGroup maps a set of concrete AWS actions to a logical access group.
type ActionGroup struct {
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}

// ConfigEntry defines a resource type and either actions or logical action groups to audit.
type ConfigEntry struct {
	ResourceType string        `json:"resource_type"`
	Actions      []string      `json:"actions,omitempty"`
	ActionGroups []ActionGroup `json:"action_groups,omitempty"`
}

func (e ConfigEntry) allActions() []string {
	if len(e.Actions) > 0 {
		return e.Actions
	}

	var actions []string
	for _, group := range e.ActionGroups {
		actions = append(actions, group.Actions...)
	}
	return actions
}

// Run executes the audit subcommand
func Run(opts *cli.Flags) {
	// Reduce GC pressure for batch workloads: the default GOGC=100 causes excessive GC when
	// many goroutines allocate concurrently, wasting ~65% of CPU time on sweep/mark/refill.
	debug.SetGCPercent(400)

	// CPU profiling via env var (CPUPROFILE=/path/to/file)
	if cpuProfile := os.Getenv("CPUPROFILE"); cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			cli.Fail("error creating CPU profile: %v", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			cli.Fail("error starting CPU profile: %v", err)
		}
		defer func() {
			pprof.StopCPUProfile()
			f.Close()
			slog.Info("wrote CPU profile", "path", cpuProfile)
		}()
	}

	if len(opts.Sources) == 0 {
		cli.Fail("error: -s/-source is required")
	}
	if opts.Config == "" {
		cli.Fail("error: -f/-config is required")
	}

	config, err := loadConfig(opts.Config)
	if err != nil {
		cli.Fail("error loading config: %v", err)
	}
	if err := validateOutputOptions(opts, config); err != nil {
		cli.Fail("error validating audit output: %v", err)
	}

	simulator, err := buildSimulator(opts.Sources)
	if err != nil {
		cli.Fail("error building simulator: %v", err)
	}

	// Build simulation options
	simOpts := []sim.OptionF{}
	if len(opts.Context) > 0 {
		simOpts = append(simOpts, sim.WithAdditionalProperties(opts.Context))
	}
	if len(opts.MultiContext) > 0 {
		simOpts = append(simOpts, sim.WithAdditionalMultiValueProperties(opts.MultiContext))
	}
	sopts := sim.NewOptions(simOpts...)

	if len(opts.OverlayFiles) > 0 {
		overlay, err := cli.LoadOverlays(opts.OverlayFiles)
		if err != nil {
			cli.Fail("error loading overlays: %v", err)
		}
		sopts.Overlay = overlay.Universe()
	}

	// Pre-freeze all principals once (reused across all config entries)
	allPrincipalArns := simulator.Universe.PrincipalArns()
	if opts.Format != formatCSV {
		sort.Strings(allPrincipalArns)
	}
	slog.Info("freezing principals", "count", len(allPrincipalArns))

	frozenPrincipals, err := simulator.FreezePrincipals(allPrincipalArns, sopts)
	if err != nil {
		cli.Fail("error freezing principals: %v", err)
	}

	// Open output
	w, err := smartrw.NewWriter(opts.Out)
	if err != nil {
		cli.Fail("error opening output: %v", err)
	}

	slog.Info("audit starting",
		"principals", len(frozenPrincipals),
		"entries", len(config),
		"format", opts.Format)

	if opts.Format == formatCSV {
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"resource", "action", "principal"}); err != nil {
			cli.Fail("error writing CSV header: %v", err)
		}

		var totalRows int
		for i, entry := range config {
			entry.Actions = entry.allActions()
			n, err := processEntry(simulator, frozenPrincipals, entry, sopts, cw)
			if err != nil {
				cli.Fail("error processing entry %d (%s): %v", i, entry.ResourceType, err)
			}
			totalRows += n
		}
		cw.Flush()
		if err := cw.Error(); err != nil {
			cli.Fail("error flushing CSV output: %v", err)
		}
		if err := w.Close(); err != nil {
			cli.Fail("error closing audit output: %v", err)
		}
		slog.Info("audit complete", "rows", totalRows)
		return
	}

	gw, err := newGroupedWriter(opts.Format, w)
	if err != nil {
		cli.Fail("error creating grouped output writer: %v", err)
	}

	var totalGroups, totalRelationships int
	for i, entry := range config {
		groups, relationships, err := processGroupedEntry(
			simulator,
			frozenPrincipals,
			entry,
			sopts,
			opts.ResourceBatchSize,
			gw,
		)
		if err != nil {
			cli.Fail("error processing entry %d (%s): %v", i, entry.ResourceType, err)
		}
		totalGroups += groups
		totalRelationships += relationships
	}
	if err := gw.Flush(); err != nil {
		cli.Fail("error flushing grouped output: %v", err)
	}
	if err := w.Close(); err != nil {
		cli.Fail("error closing audit output: %v", err)
	}

	slog.Info("audit complete", "groups", totalGroups, "relationships", totalRelationships)
}

// loadConfig reads and parses the audit config JSON file
func loadConfig(path string) ([]ConfigEntry, error) {
	reader, err := smartrw.NewReader(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open config: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to read config: %w", err)
	}

	var config []ConfigEntry
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	for i := range config {
		entry := &config[i]
		if entry.ResourceType == "" {
			return nil, fmt.Errorf("entry %d: missing resource_type", i)
		}
		if (len(entry.Actions) == 0) == (len(entry.ActionGroups) == 0) {
			return nil, fmt.Errorf(
				"entry %d (%s): must supply exactly one of actions or action_groups",
				i,
				entry.ResourceType,
			)
		}

		if len(entry.Actions) > 0 {
			for j, action := range entry.Actions {
				canonical, err := canonicalAction(action)
				if err != nil {
					return nil, fmt.Errorf("entry %d (%s): %w", i, entry.ResourceType, err)
				}
				entry.Actions[j] = canonical
			}
			continue
		}

		groupNames := make(map[string]struct{})
		actionNames := make(map[string]string)
		for j := range entry.ActionGroups {
			group := &entry.ActionGroups[j]
			if group.Name == "" {
				return nil, fmt.Errorf(
					"entry %d (%s): action group %d is missing name",
					i,
					entry.ResourceType,
					j,
				)
			}
			if len(group.Actions) == 0 {
				return nil, fmt.Errorf(
					"entry %d (%s): action group %q is missing actions",
					i,
					entry.ResourceType,
					group.Name,
				)
			}
			if _, ok := groupNames[group.Name]; ok {
				return nil, fmt.Errorf(
					"entry %d (%s): duplicate action group %q",
					i,
					entry.ResourceType,
					group.Name,
				)
			}
			groupNames[group.Name] = struct{}{}

			for k, action := range group.Actions {
				canonical, err := canonicalAction(action)
				if err != nil {
					return nil, fmt.Errorf("entry %d (%s): %w", i, entry.ResourceType, err)
				}
				if previousGroup, ok := actionNames[canonical]; ok {
					if previousGroup == group.Name {
						return nil, fmt.Errorf(
							"entry %d (%s): duplicate action %q in action group %q",
							i,
							entry.ResourceType,
							canonical,
							group.Name,
						)
					}
					return nil, fmt.Errorf(
						"entry %d (%s): action %q belongs to both %q and %q",
						i,
						entry.ResourceType,
						canonical,
						previousGroup,
						group.Name,
					)
				}
				actionNames[canonical] = group.Name
				group.Actions[k] = canonical
			}
		}
	}

	return config, nil
}

func canonicalAction(action string) (string, error) {
	resolved, ok := sar.LookupString(action)
	if !ok {
		return "", fmt.Errorf("unknown action %q", action)
	}
	return resolved.ShortName(), nil
}

func validateOutputOptions(opts *cli.Flags, config []ConfigEntry) error {
	switch opts.Format {
	case formatCSV:
		return nil
	case formatGroupedCSV, formatGroupedJSONL:
		if opts.ResourceBatchSize <= 0 {
			return fmt.Errorf("resource-batch-size must be greater than zero")
		}
		for i, entry := range config {
			if len(entry.ActionGroups) == 0 {
				return fmt.Errorf(
					"entry %d (%s): %s output requires action_groups",
					i,
					entry.ResourceType,
					opts.Format,
				)
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported format %q", opts.Format)
	}
}

// buildSimulator creates a Simulator with data loaded from the specified sources
func buildSimulator(sources []string) (*sim.Simulator, error) {
	simulator, err := sim.NewSimulator()
	if err != nil {
		return nil, fmt.Errorf("unable to create simulator: %w", err)
	}

	for _, src := range sources {
		reader, err := smartrw.NewReader(src)
		if err != nil {
			return nil, fmt.Errorf("unable to open source '%s': %w", src, err)
		}

		source := server.Source{Reader: *reader}
		uv, err := source.Universe()
		if err != nil {
			return nil, fmt.Errorf("unable to load source '%s': %w", src, err)
		}

		simulator.Universe.Merge(uv)
		slog.Info("loaded source", "source", src, "size", simulator.Universe.Size())
	}

	return simulator, nil
}

// processEntry runs the audit for a single config entry, streaming results directly to CSV
func processEntry(
	simulator *sim.Simulator,
	frozenPrincipals []*entities.FrozenPrincipal,
	entry ConfigEntry,
	opts sim.Options,
	cw *csv.Writer,
) (int, error) {
	// Filter resources by type
	var resourceArns []string
	for r := range simulator.Universe.Resources() {
		if r.Type == entry.ResourceType {
			resourceArns = append(resourceArns, r.Arn)
		}
	}

	if len(resourceArns) == 0 {
		slog.Info("no resources found for type", "type", entry.ResourceType)
		return 0, nil
	}

	// Expand resources (e.g. S3 bucket -> object)
	expanded, err := simulator.ExpandResources(resourceArns, opts)
	if err != nil {
		return 0, fmt.Errorf("unable to expand resources: %w", err)
	}

	// Freeze resources for this entry
	frozenResources, err := simulator.FreezeResources(expanded, opts)
	if err != nil {
		return 0, fmt.Errorf("unable to freeze resources: %w", err)
	}

	slog.Info("processing entry",
		"type", entry.ResourceType,
		"resources", len(frozenResources),
		"actions", len(entry.Actions),
		"principals", len(frozenPrincipals))

	// Stream results: dedup and write CSV inline instead of collecting all in memory
	seen := make(map[string]struct{})
	var rows int
	var writeErr error

	err = simulator.ProductFrozenStreaming(
		frozenPrincipals,
		entry.Actions,
		frozenResources,
		opts,
		func(t sim.AccessTuple) {
			if writeErr != nil {
				return
			}

			resource := collapseS3Arn(t.Resource)
			key := resource + "\x00" + t.Action + "\x00" + t.Principal
			if _, ok := seen[key]; ok {
				return
			}
			seen[key] = struct{}{}

			if err := cw.Write([]string{resource, t.Action, t.Principal}); err != nil {
				writeErr = fmt.Errorf("error writing CSV row: %w", err)
				return
			}
			rows++
		},
	)
	if err != nil {
		return 0, fmt.Errorf("simulation error: %w", err)
	}
	if writeErr != nil {
		return 0, writeErr
	}

	slog.Info("entry complete",
		"type", entry.ResourceType,
		"allowed", rows)

	return rows, nil
}

// collapseS3Arn strips the object path from S3 object ARNs back to the bucket
func collapseS3Arn(arn string) string {
	if strings.HasPrefix(arn, "arn:aws:s3:::") && strings.Contains(arn, "/") {
		return strings.SplitN(arn, "/", 2)[0]
	}
	return arn
}
