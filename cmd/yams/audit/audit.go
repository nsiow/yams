package audit

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/internal/smartrw"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/server"
	"github.com/nsiow/yams/pkg/sim"
)

// ConfigEntry defines a resource type and the actions to audit against it
type ConfigEntry struct {
	ResourceType string   `json:"resource_type"`
	Actions      []string `json:"actions"`
}

// Run executes the audit subcommand
func Run(opts *cli.Flags) {
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

	simulator, err := buildSimulator(opts.Sources)
	if err != nil {
		cli.Fail("error building simulator: %v", err)
	}

	// Build simulation options
	simOpts := []sim.OptionF{}
	if len(opts.Context) > 0 {
		simOpts = append(simOpts, sim.WithAdditionalProperties(opts.Context))
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
	defer w.Close()

	cw := csv.NewWriter(w)
	defer cw.Flush()

	if err := cw.Write([]string{"resource", "action", "principal"}); err != nil {
		cli.Fail("error writing CSV header: %v", err)
	}

	slog.Info("audit starting",
		"principals", len(frozenPrincipals),
		"entries", len(config))

	var totalRows int
	for i, entry := range config {
		n, err := processEntry(simulator, frozenPrincipals, entry, sopts, cw)
		if err != nil {
			cli.Fail("error processing entry %d (%s): %v", i, entry.ResourceType, err)
		}
		totalRows += n
	}

	slog.Info("audit complete", "rows", totalRows)
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

	for i, entry := range config {
		if entry.ResourceType == "" {
			return nil, fmt.Errorf("entry %d: missing resource_type", i)
		}
		if len(entry.Actions) == 0 {
			return nil, fmt.Errorf("entry %d (%s): missing actions", i, entry.ResourceType)
		}
	}

	return config, nil
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
