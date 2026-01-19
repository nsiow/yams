package cli

import (
	"log/slog"
	"os"
	"strings"
)

// Verbose is set when --verbose or -V flag is passed
var Verbose bool

func InitLogging() {
	var logLevel slog.Level

	// Check env var first
	env := strings.ToLower(os.Getenv("YAMS_DEBUG"))
	if env == "1" || env == "true" || env == "yes" || env == "on" {
		logLevel = slog.LevelDebug
	}

	// Check for --verbose/-V flag before subcommand, remove if found
	var newArgs []string
	for i, arg := range os.Args {
		if arg == "--verbose" || arg == "-V" {
			Verbose = true
			logLevel = slog.LevelDebug
		} else {
			newArgs = append(newArgs, os.Args[i])
		}
	}
	os.Args = newArgs

	// Create handler with text output to stderr (keeps stdout clean for JSON output)
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})

	// Create and set default logger
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
