package main

import "log/slog"

func main() {
	// Set up CLI logging
	initLogging()

	// Parse CLI arguments
	flags, err := ParseFlags()
	if err != nil {
		exit("error: %v", err)
	}
	slog.Debug("cli flags", "flags", flags)

	// Run the requested command logic
	switch flags.Mode {
	case RUN_MODE_DUMP:
		runDump(flags)
	}
}
