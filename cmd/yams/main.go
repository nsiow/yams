package main

import (
	"fmt"
	"os"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
)

func main() {
	// Parse CLI arguments
	rc, err := ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cli error: %v", err)
		os.Exit(1)
	}
	IS_DEBUG_ENABLED = rc.Debug
	debug("running %v with flags: %+v", os.Args[0], rc)

	// Read the provided cache file
	data, err := os.ReadFile(rc.Cache)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read cache file: %v", err)
		os.Exit(1)
	}

	// Attempt to parse the data
	var env entities.Environment
	switch rc.CacheFormat {
	case CONST_CACHE_FORMAT_AWS_CONFIG:
		l := awsconfig.NewLoader()
		err = l.LoadJson(data)
		env = l.Environment()
	case CONST_CACHE_FORMAT_AWS_CONFIG_LINES:
		l := awsconfig.NewLoader()
		err = l.LoadJsonl(data)
		env = l.Environment()
	default:
		fmt.Fprintf(os.Stderr, "unsure how to load cache format: %s", rc.CacheFormat)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to load cache: %v", err)
		os.Exit(1)
	}
	fmt.Printf("loaded %d principals; %d resources from cache\n",
		len(env.Principals), len(env.Resources))
}
