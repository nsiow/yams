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

func PrintVersion() {
	fmt.Printf("yams %s\n", Version)
	fmt.Printf("  commit:  %s\n", GitCommit)
	fmt.Printf("  built:   %s\n", BuildDate)
	fmt.Printf("  go:      %s\n", runtime.Version())
}
