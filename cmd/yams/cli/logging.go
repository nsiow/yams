package cli

import (
	"log/slog"
	"os"
	"strings"
)

func InitLogging() {

	var logLevel slog.Level

	env := strings.ToLower(os.Getenv("YAMS_DEBUG"))
	if env == "1" || env == "true" || env == "yes" || env == "on" {
		logLevel = slog.LevelDebug
	}

	// Create handler with JSON output to stdout
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	// Create and set default logger
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
