package main

import (
	"fmt"
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

	switch opts.DumpTarget {
	case DUMP_TARGET_ORG:
		output, err = dumpOrg()
		if err != nil {
			exit("error attempt to dump org data: %v", err.Error())
		}
	default:
		exit("unknown dump target '%s', must be one of: %s", opts.DumpTarget, DUMP_TARGETS)
	}

	fmt.Println(output)
}
