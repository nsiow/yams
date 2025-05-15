package log

import (
	"log/slog"
	"os"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestBasicLogger(t *testing.T) {
	l := Logger("foo")
	l.Info("message", "foo", "bar")
}

func TestConvertLevel(t *testing.T) {
	// Save original value and reset at end of function
	orig := os.Getenv(YAMS_LOG_LEVEL_ENV_VAR)
	defer func() { os.Setenv(YAMS_LOG_LEVEL_ENV_VAR, orig) }()

	tests := []testlib.TestCase[string, slog.Level]{
		{
			Input: `ERROR`,
			Want:  slog.LevelError,
		},
		{
			Input: `warn`,
			Want:  slog.LevelWarn,
		},
		{
			Input: `iNfO`,
			Want:  slog.LevelInfo,
		},
		{
			Input: `DEBUG`,
			Want:  slog.LevelDebug,
		},
		{
			Input: `anything-else`,
			Want:  nothing,
		},
		{
			Input: ``,
			Want:  nothing,
		},
	}

	testlib.RunTestSuite(t, tests, func(i string) (slog.Level, error) {
		return convertLevel(i), nil
	})
}
