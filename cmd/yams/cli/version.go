package cli

import (
	"fmt"
	"runtime"
)

// Version info, populated via ldflags at build time
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// DefaultServerAddress is the default -server address used by client subcommands
// (status, sim, inventory, etc.). It can be overridden at build time via:
//
//	go build -ldflags "-X github.com/nsiow/yams/cmd/yams/cli.DefaultServerAddress=..."
//
// The YAMS_SERVER_ADDRESS env var still takes precedence at runtime.
var DefaultServerAddress = ":8888"

func PrintVersion() {
	fmt.Printf("yams %s\n", Version)
	fmt.Printf("  commit:  %s\n", GitCommit)
	fmt.Printf("  built:   %s\n", BuildDate)
	fmt.Printf("  go:      %s\n", runtime.Version())
}
