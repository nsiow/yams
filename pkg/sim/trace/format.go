package trace

import "fmt"

// format is a helper function for sprintf-style formatting of records
//
// This is factored out for two reasons:
// - avoid allocations for arg-less formatting calls
// - provide a single place for any formatting customizations
func format(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}

	return fmt.Sprintf(msg, args...)
}
