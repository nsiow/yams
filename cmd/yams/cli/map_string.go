package cli

import (
	"fmt"
	"strings"

	json "github.com/bytedance/sonic"
)

// MapString implements the flag.Value interface for key/value pairs specified via the CLI.
// It accepts both key=value format and JSON object format: {"key": "value", ...}
type MapString map[string]string

func (m *MapString) String() string {
	basicMap := map[string]string(*m)
	return fmt.Sprintf("%+v", basicMap)
}

func (m *MapString) Set(value string) error {
	if len(*m) == 0 {
		*m = make(MapString)
	}

	// Check if value looks like JSON
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		var jsonMap map[string]string
		if err := json.Unmarshal([]byte(trimmed), &jsonMap); err != nil {
			return fmt.Errorf("invalid JSON for context: %w", err)
		}
		for k, v := range jsonMap {
			(*m)[k] = v
		}
		return nil
	}

	// Fall back to key=value format
	substr := strings.SplitN(value, "=", 2)
	if len(substr) != 2 {
		return fmt.Errorf("unable to split k/v pairs for MapString: %s", value)
	}

	(*m)[substr[0]] = substr[1]
	return nil
}
