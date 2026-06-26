package cli

import (
	"fmt"
	"strings"

	json "github.com/bytedance/sonic"
)

// MapStringList implements the flag.Value interface for key/list pairs specified via the CLI.
type MapStringList map[string][]string

func (m *MapStringList) String() string {
	basicMap := map[string][]string(*m)
	return fmt.Sprintf("%+v", basicMap)
}

func (m *MapStringList) Set(value string) error {
	if len(*m) == 0 {
		*m = make(MapStringList)
	}

	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		var jsonMap map[string][]string
		if err := json.Unmarshal([]byte(trimmed), &jsonMap); err != nil {
			return fmt.Errorf("invalid JSON for multi-value context: %w", err)
		}
		for k, v := range jsonMap {
			(*m)[k] = v
		}
		return nil
	}

	substr := strings.SplitN(value, "=", 2)
	if len(substr) != 2 {
		return fmt.Errorf("unable to split k/v pairs for MapStringList: %s", value)
	}

	(*m)[substr[0]] = append((*m)[substr[0]], substr[1])
	return nil
}
