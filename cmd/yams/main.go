package main

import (
	"fmt"
	"os"
)

func main() {
	// Parse CLI arguments
	flags, err := ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(2)
	}

	// Run the requested command logic
	switch flags.mode {
	case RUN_MODE_DUMP:
		runDump(flags)
	}
}
