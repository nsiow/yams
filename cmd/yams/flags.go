package main

import (
	"flag"
	"fmt"
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
	mode string

	// dump
	Target string

	// server
	Cache       string
	CacheFormat string
	Debug       bool
}

func ParseFlags() (*Flags, error) {
	// Define empty run command; we'll aim to mostly use flag.*Var
	opts := Flags{}

	// Check for subcommand
	if len(os.Args) < 2 || !slices.Contains(VALID_RUN_MODES, os.Args[1]) {
		return nil, fmt.Errorf(
			"invalid command for %s: must provide one of %v\n", os.Args[0], VALID_RUN_MODES)
	}
	opts.mode = os.Args[1]

	// Parse options specific to subcommand
	switch opts.mode {
	case RUN_MODE_DUMP:
		initFlagsForDump(&opts)
	}

	// Populate + validate flags, then return runconfig
	flag.Parse()
	return &opts, ValidateFlags(opts)
}

func ValidateFlags(rc Flags) error {
	// Make sure a cache file was provided
	// if rc.Cache == "" {
	// 	return fmt.Errorf("no resource cache found; provide one with '-cache'")
	// }

	return nil
}
