package main

import (
	"fmt"
	"os"
	"slices"

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

	// Read the provided cache file
	file, err := os.Open(rc.Cache)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open cache file: %v", err)
		os.Exit(1)
	}

	// Attempt to parse the data
	var uv *entities.Universe
	switch rc.CacheFormat {
	case CONST_CACHE_FORMAT_AWS_CONFIG:
		loader := awsconfig.NewLoader()
		err = loader.LoadJson(file)
		uv = loader.Universe()
	case CONST_CACHE_FORMAT_AWS_CONFIG_LINES:
		loader := awsconfig.NewLoader()
		err = loader.LoadJsonl(file)
		uv = loader.Universe()
	default:
		fmt.Fprintf(os.Stderr, "unsure how to load cache format: %s", rc.CacheFormat)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to load cache: %v", err)
		os.Exit(1)
	}
	fmt.Printf("loaded %d principals; %d resources from cache\n",
		len(slices.Collect(uv.Principals())),
		len(slices.Collect(uv.Resources())))
}
