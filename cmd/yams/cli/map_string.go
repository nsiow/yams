package cli

import (
	"fmt"
	"strings"
)

// MapString implements the flag.Value interface for key/value pairs specified via the CLI.
type MapString map[string]string

func (m *MapString) String() string {
	basicMap := map[string]string(*m)
	return fmt.Sprintf("%+v", basicMap)
}

func (m *MapString) Set(value string) error {
	if len(*m) == 0 {
		*m = make(MapString)
	}

	substr := strings.SplitN(value, "=", 2)
	if len(substr) != 2 {
		return fmt.Errorf("unable to split k/v pairs for MapString: %s", value)
	}

	(*m)[substr[0]] = substr[1]
	return nil
}
