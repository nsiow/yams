package cli

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	v1 "github.com/nsiow/yams/pkg/server/api/v1"
)

const (
	RUN_MODE_STATUS     = "status"
	RUN_MODE_DUMP       = "dump"
	RUN_MODE_SERVER     = "server"
	RUN_MODE_ACCOUNTS   = "accounts"
	RUN_MODE_ACTIONS    = "actions"
	RUN_MODE_RESOURCES  = "resources"
	RUN_MODE_PRINCIPALS = "principals"
	RUN_MODE_POLICIES   = "policies"
	RUN_MODE_SIM        = "sim"
)

var RUN_MODES = []string{
	RUN_MODE_STATUS,
	RUN_MODE_DUMP,
	RUN_MODE_SERVER,
	RUN_MODE_ACCOUNTS,
	RUN_MODE_ACTIONS,
	RUN_MODE_RESOURCES,
	RUN_MODE_PRINCIPALS,
	RUN_MODE_POLICIES,
	RUN_MODE_SIM,
}

// Flags is a struct containing all flags/options related to CLI behavior
type Flags struct {
	Mode string

	// dump
	Target        string
	Out           string
	Aggregator    string
	ResourceTypes MultiString
	DryRun        bool

	// server
	Addr         string
	Sources      MultiString
	Refresh      int
	Debug        bool
	Env          MultiString
	OverlayStore string

	// inventory
	Key    string
	Query  string
	Freeze bool
	Format string

	// sim
	Principal    string
	Action       string
	Resource     string
	Context      MapString
	Explain      bool
	Trace        bool
	OverlayFiles MultiString
	Overlay      v1.Overlay
	Exact        bool

	// multiple
	Server string
}

func Parse() (*Flags, error) {
	// Define empty run command; we'll aim to mostly use flag.*Var
	opts := &Flags{}
	var args []string
	var err error

	// Handle version and completion commands early
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "version", "-v", "--version":
			PrintVersion()
			os.Exit(0)
		case "completion":
			shell := "bash"
			if len(os.Args) >= 3 {
				shell = os.Args[2]
			}
			PrintCompletion(shell)
			os.Exit(0)
		}
	}

	// Check for subcommand
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		PrintHelp()
		os.Exit(0)
	}

	// Resolve command alias
	opts.Mode = ResolveAlias(os.Args[1])
	slog.Debug("parsed mode", "mode", opts.Mode)

	// Parse options specific to subcommand
	switch opts.Mode {

	case RUN_MODE_STATUS:
		fs := flag.NewFlagSet("status", flag.ExitOnError)

		fs.StringVar(&opts.Server, "s", "", "alias for -server")
		fs.StringVar(&opts.Server, "server", ":8888", "address of yams server to use for connection")

		fs.StringVar(&opts.Format, "format", "json", "output format: json or table")

		err = fs.Parse(os.Args[2:])
		args = fs.Args()

	case RUN_MODE_DUMP:
		fs := flag.NewFlagSet("dump", flag.ExitOnError)

		fs.StringVar(&opts.Target, "t", "", "alias for -target")
		fs.StringVar(&opts.Target, "target", "", "which target to dump, one of: [config, org]")

		fs.StringVar(&opts.Out, "o", "", "alias for -out")
		fs.StringVar(&opts.Out, "out", "",
			"destination target for writing, such as out.json or file:///tmp/out.json")

		fs.StringVar(&opts.Aggregator, "a", "", "alias for -aggregator")
		fs.StringVar(&opts.Aggregator, "aggregator", "", "name of the AWS Config aggregator to use")

		fs.Var(&opts.ResourceTypes, "r", "alias for -rtype")
		fs.Var(&opts.ResourceTypes, "rtype", "resource type(s) to dump, e.g. 'AWS::SQS::Queue'")

		fs.BoolVar(&opts.DryRun, "n", false, "alias for -dry-run")
		fs.BoolVar(&opts.DryRun, "dry-run", false, "show what would be done without executing")

		err = fs.Parse(os.Args[2:])
		args = fs.Args()

	case RUN_MODE_SERVER:
		fs := flag.NewFlagSet("server", flag.ExitOnError)

		fs.StringVar(&opts.Addr, "a", ":8888", "alias for -addr")
		fs.StringVar(&opts.Addr, "addr", ":8888", "address for running server")

		fs.Var(&opts.Sources, "s", "alias for -source")
		fs.Var(&opts.Sources, "source", "list of sources to use for server data (supports multiple)")

		fs.IntVar(&opts.Refresh, "r", 0, "alias for -refresh")
		fs.IntVar(&opts.Refresh, "refresh", 0,
			"refresh rate (in seconds) for specified sources; defaults to no refresh")

		fs.Var(&opts.Env, "e", "alias for -env")
		fs.Var(&opts.Env, "env", "environment variables to report in /status endpoint")

		fs.StringVar(&opts.OverlayStore, "overlay", "",
			"overlay store backend: 'memory' (default) or 'ddb://<table-name>'")

		err = fs.Parse(os.Args[2:])
		args = fs.Args()

	case
		RUN_MODE_ACCOUNTS,
		RUN_MODE_ACTIONS,
		RUN_MODE_POLICIES,
		RUN_MODE_PRINCIPALS,
		RUN_MODE_RESOURCES:

		subcommands := []string{
			RUN_MODE_ACCOUNTS,
			RUN_MODE_ACTIONS,
			RUN_MODE_POLICIES,
			RUN_MODE_PRINCIPALS,
			RUN_MODE_RESOURCES,
		}

		var fs *flag.FlagSet
		for _, subcommand := range subcommands {
			fs = flag.NewFlagSet(subcommand, flag.ExitOnError)

			fs.StringVar(&opts.Server, "s", "", "alias for -server")
			fs.StringVar(&opts.Server, "server", ":8888", "address of yams server to use for connection")

			fs.StringVar(&opts.Query, "q", "", "alias for -query")
			fs.StringVar(&opts.Query, "query", "", "case-insensitive search term")

			fs.StringVar(&opts.Key, "k", "", "alias for -key")
			fs.StringVar(&opts.Key, "key", "", "primary key of requested entity (ARN, Account ID, etc)")

			fs.BoolVar(&opts.Freeze, "f", false, "alias for -freeze")
			fs.BoolVar(&opts.Freeze, "freeze", false,
				"freeze the entity if applicable, resolving all references to a snapshotted state")

			fs.StringVar(&opts.Format, "format", "json", "output format: json or table")
		}

		err = fs.Parse(os.Args[2:])
		args = fs.Args()

	case RUN_MODE_SIM:
		fs := flag.NewFlagSet("sim", flag.ExitOnError)

		fs.StringVar(&opts.Server, "s", "", "alias for -server")
		fs.StringVar(&opts.Server, "server", ":8888", "address of yams server to use for connection")

		fs.StringVar(&opts.Principal, "p", "", "alias for -principal")
		fs.StringVar(&opts.Principal, "principal", "", "ARN of the Principal to simulate")

		fs.StringVar(&opts.Action, "a", "", "alias for -action")
		fs.StringVar(&opts.Action, "action", "", "AWS API call to simulate")

		fs.StringVar(&opts.Resource, "r", "", "alias for -resource")
		fs.StringVar(&opts.Resource, "resource", "", "ARN of the Resource to simulate")

		fs.Var(&opts.Context, "c", "alias for -context")
		fs.Var(&opts.Context, "context", "Additional request-context property for simulation")

		fs.Var(&opts.OverlayFiles, "o", "alias for -overlay")
		fs.Var(&opts.OverlayFiles, "overlay", "Entity definition file for overrides")

		fs.BoolVar(&opts.Exact, "x", false, "alias for -exact")
		fs.BoolVar(&opts.Exact, "exact", false, "disable fuzzy-matching for ARNs")

		fs.BoolVar(&opts.Explain, "e", false, "alias for -explain")
		fs.BoolVar(&opts.Explain, "explain", false,
			"provide additional context on how the decision was reached")

		fs.BoolVar(&opts.Trace, "t", false, "alias for -trace")
		fs.BoolVar(&opts.Trace, "trace", false,
			"provide full evaluation context on how the decision was reached")

		err = fs.Parse(os.Args[2:])
		args = fs.Args()

	// unknown mode
	default:
		return nil, fmt.Errorf("'%s' is not one of available commands: %s",
			opts.Mode, strings.Join(RUN_MODES, ", "))
	}
	slog.Debug("opts after flag parsing", "opts", opts)

	if len(args) > 0 {
		return nil, fmt.Errorf("unknown argument: %s", args[0])
	}

	// Apply config file defaults (only if values weren't explicitly set)
	if cfg := LoadConfig(); cfg != nil {
		if opts.Server == "" || opts.Server == ":8888" {
			if cfg.Server != "" {
				opts.Server = cfg.Server
			}
		}
		if opts.Format == "" || opts.Format == "json" {
			if cfg.Format != "" {
				opts.Format = cfg.Format
			}
		}
	}

	// Allow address override via environment (highest priority)
	envserver := os.Getenv("YAMS_SERVER_ADDRESS")
	if len(envserver) > 0 {
		opts.Server = envserver
	}

	return opts, err
}
