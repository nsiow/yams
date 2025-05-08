package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"
)

const (
	RUN_MODE_DUMP   = "dump"
	RUN_MODE_SERVER = "server"
)

var VALID_RUN_MODES = []string{
	RUN_MODE_DUMP,
	RUN_MODE_SERVER,
}

// Flags is a struct containing all flags/options related to CLI behavior
type Flags struct {
	Mode string

	// dump
	DumpTarget string

	// server
	Cache string
	Debug bool
}

func ParseFlags() (*Flags, error) {
	// Define empty run command; we'll aim to mostly use flag.*Var
	opts := &Flags{}

	// Check for subcommand
	if len(os.Args) < 2 || !slices.Contains(VALID_RUN_MODES, os.Args[1]) {
		return nil, fmt.Errorf(
			"invalid command for %s: must provide one of %v\n", os.Args[0], VALID_RUN_MODES)
	}
	opts.Mode = os.Args[1]
	slog.Debug("parsed mode", "mode", opts.Mode)

	// Parse options specific to subcommand
	switch opts.Mode {

	// mode=dump
	case RUN_MODE_DUMP:
		fs := flag.NewFlagSet("dump", flag.ExitOnError)
		fs.StringVar(&opts.DumpTarget, "target", opts.DumpTarget,
			fmt.Sprintf("which target to dump, one of: %v", DUMP_TARGETS))
		fs.Parse(os.Args[2:])

	// should never get here
	default:
		panic("invalid unreachable mode somehow?")
	}
	slog.Debug("opts after flag parsing", "opts", opts)

	return opts, nil
}
