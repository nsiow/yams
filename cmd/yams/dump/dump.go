package dump

import (
	"fmt"
	"os"

	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/internal/smartrw"
)

// Logic for the "dump" subcommand
func Run(opts *cli.Flags) {
	// Validate target first
	switch opts.Target {
	case "org", "config":
		// valid
	case "":
		cli.Fail("missing required flag: -t/--target")
	default:
		cli.Fail("unknown dump target '%s'", opts.Target)
	}

	// Handle dry-run mode
	if opts.DryRun {
		printDryRun(opts)
		return
	}

	var output string
	var err error

	writer, err := smartrw.NewWriter(opts.Out)
	if err != nil {
		cli.Fail("invalid destination '%s': %v", opts.Out, err)
	}

	switch opts.Target {
	case "org":
		output, err = Org(opts)
	case "config":
		output, err = Config(opts)
	}

	if err != nil {
		cli.Fail("error attempting to dump '%s': %v", opts.Target, err)
	}

	_, err = writer.Write([]byte(output))
	if err != nil {
		cli.Fail("error writing to destination '%s': %v", opts.Out, err)
	}

	err = writer.Close()
	if err != nil {
		cli.Fail("error closing/flushing to destination '%s': %v", opts.Out, err)
	}
}

func printDryRun(opts *cli.Flags) {
	fmt.Fprintln(os.Stderr, "Dry run mode - no changes will be made")
	fmt.Fprintln(os.Stderr)

	dest := opts.Out
	if dest == "" {
		dest = "stdout"
	}

	switch opts.Target {
	case "org":
		fmt.Fprintln(os.Stderr, "Would dump AWS Organizations data:")
		fmt.Fprintln(os.Stderr, "  - Organization accounts")
		fmt.Fprintln(os.Stderr, "  - Service Control Policies (SCPs)")
		fmt.Fprintln(os.Stderr, "  - Resource Control Policies (RCPs)")
		fmt.Fprintf(os.Stderr, "  - Output destination: %s\n", dest)
	case "config":
		fmt.Fprintln(os.Stderr, "Would dump AWS Config data:")
		fmt.Fprintf(os.Stderr, "  - Aggregator: %s\n", opts.Aggregator)
		if len(opts.ResourceTypes) > 0 {
			fmt.Fprintf(os.Stderr, "  - Resource types: %v\n", []string(opts.ResourceTypes))
		} else {
			fmt.Fprintln(os.Stderr, "  - Resource types: all")
		}
		fmt.Fprintf(os.Stderr, "  - Output destination: %s\n", dest)
	}
}
