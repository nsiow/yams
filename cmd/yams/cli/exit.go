package cli

import (
	"fmt"
	"os"
	"strings"

	json "github.com/bytedance/sonic"
)

// Common error hints based on error message patterns
var errorHints = map[string]string{
	"connection refused": "Is the yams server running? Start it with: yams server -s <source>",
	"no such host":       "Check the server address. Use -s/--server or set YAMS_SERVER_ADDRESS",
	"timeout":            "Server may be overloaded or unreachable. Check network connectivity",
	"unknown command":    "Run 'yams -h' to see available commands",
	"unknown mode":       "Run 'yams -h' to see available commands",
}

// getHint returns a helpful hint for the given error message
func getHint(errMsg string) string {
	lower := strings.ToLower(errMsg)
	for pattern, hint := range errorHints {
		if strings.Contains(lower, pattern) {
			return hint
		}
	}
	return ""
}

func Fail(format string, a ...any) {
	errMsg := fmt.Sprintf(format, a...)

	if StderrIsTTY() {
		// Human-friendly output for TTY
		fmt.Fprintf(os.Stderr, "error: %s\n", errMsg)
		if hint := getHint(errMsg); hint != "" {
			fmt.Fprintf(os.Stderr, "hint: %s\n", hint)
		}
	} else {
		// JSON output for non-TTY (piped/scripted)
		errblob, err := json.MarshalIndent(
			map[string]string{
				"error": errMsg,
			},
			"",
			"  ",
		)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(os.Stderr, "%s\n", errblob)
	}

	os.Exit(2)
}
