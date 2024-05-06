package main

import (
	"flag"
	"fmt"
	"slices"
	"strings"
)

// RunConfig is a struct containing all flags/options related to CLI behavior
type RunConfig struct {
	Cache       string
	CacheFormat string
	Debug       bool
}

func ParseFlags() (RunConfig, error) {
	// Define empty run command; we'll aim to mostly use flag.*Var
	rc := RunConfig{}

	// Define interface
	flag.StringVar(&rc.Cache, "cache", "",
		"cachefile for resources")
	flag.StringVar(&rc.CacheFormat, "cacheformat", "",
		fmt.Sprintf("format of the cache, one of %v", CONST_CACHE_FORMAT_OPTIONS))
	flag.BoolVar(&rc.Debug, "debug", false,
		"enable enhanced debugging behavior")

	// Populate + validate flags, then return runconfig
	flag.Parse()
	return rc, ValidateFlags(rc)
}

func ValidateFlags(rc RunConfig) error {
	// Make sure a cache file was provided
	if rc.Cache == "" {
		return fmt.Errorf("no resource cache found; provide one with '-cache'")
	}

	// If -cacheformat was not provided, try to infer
	if rc.CacheFormat == "" {
		if strings.HasSuffix(rc.Cache, ".awsconfig.json") {
			rc.CacheFormat = CONST_CACHE_FORMAT_AWS_CONFIG
		} else if strings.HasSuffix(rc.Cache, ".awsconfig.jsonl") {
			rc.CacheFormat = CONST_CACHE_FORMAT_AWS_CONFIG_LINES
		}
	}

	// If -cacheformat is invalid, throw an error
	if !slices.Contains(CONST_CACHE_FORMAT_OPTIONS, rc.CacheFormat) {
		return fmt.Errorf("invalid -cacheformat '%s', should be one of %v",
			rc.CacheFormat, CONST_CACHE_FORMAT_OPTIONS)
	}

	return nil
}
