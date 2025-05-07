package main

import (
	"flag"
	"fmt"
)

// DUMP_TARGETS defines data sets that are valid targets of the "dump" subcommand
var DUMP_TARGETS = []string{
	"org",
}

// Define all CLI options for the "dump" subcommand
func initFlagsForDump(opts *Flags) {
	flag.StringVar(&opts.Target, "target", "",
		fmt.Sprintf("which target to dump, one of: %v", DUMP_TARGETS))
}

// Logic for the "dump" subcommand
func runDump(opts *Flags) {
}
