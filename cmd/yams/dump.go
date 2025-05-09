package main

import (
	"slices"

	"github.com/nsiow/yams/internal/smartrw"
)

const (
	DUMP_TARGET_ORG = "org"
)

// DUMP_TARGETS defines data sets that are valid targets of the "dump" subcommand
var DUMP_TARGETS = []string{
	DUMP_TARGET_ORG,
}

// Logic for the "dump" subcommand
func runDump(opts *Flags) {
	var output string
	var err error

	writer, err := smartrw.NewWriter(opts.OutFile)
	if err != nil {
		fail("error opening destination '%s' for writing: %v", opts.OutFile, err)
	}

	if !slices.Contains(DUMP_TARGETS, opts.DumpTarget) {
		fail("unknown dump target '%s', must be one of: %s", opts.DumpTarget, DUMP_TARGETS)
	}

	if opts.DumpTarget == DUMP_TARGET_ORG {
		output, err = dumpOrg()
	}

	if err != nil {
		fail("error attempting to dump '%s': %v", opts.DumpTarget, err)
	}

	_, err = writer.Write([]byte(output))
	if err != nil {
		fail("error writing to destination '%s': %v", opts.OutFile, err)
	}

	err = writer.Close()
	if err != nil {
		fail("error closing/flushing to destination '%s': %v", opts.OutFile, err)
	}
}
