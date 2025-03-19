package log

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

// NOTHING represents a custom logging level for absolutely no logs
var nothing slog.Level = 99

// logger is the shared logger
var logger *slog.Logger

// loggerOnce is a do-once mutex ensuring we only ever create 1 logger instance
var loggerOnce sync.Once

// Logger returns a logger with the name preconfigured to logging attributes
func Logger(name string) *slog.Logger {
	loggerOnce.Do(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLevel(),
		}))
	})

	return logger.With("logger", name)
}

// getLevel returns the correct logging level based on environment variables
func getLevel() slog.Level {
	switch strings.ToUpper(os.Getenv("YAMS_LOG_LEVEL")) {
	case "ERROR":
		return slog.LevelError
	case "WARN":
		return slog.LevelWarn
	case "INFO":
		return slog.LevelInfo
	case "DEBUG":
		return slog.LevelDebug
	default:
		return nothing
	}
}
