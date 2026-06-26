package policy

import (
	stdjson "encoding/json"
	"fmt"
	"slices"
	"strings"

	json "github.com/bytedance/sonic"
)

// Value is a JSON-centric helper struct to facilitate one-or-more value representations
type Value []string

// NewValue creates a new PolicyString struct using the supplied values
func NewValue(values ...string) Value {
	return values
}

// MarshalJSON instructs how to convert Value fields to raw bytes
func (v Value) MarshalJSON() ([]byte, error) {
	switch v.Count() {
	case 1:
		return json.Marshal(v[0])
	default:
		return json.Marshal([]string(v))
	}
}

// UnmarshalJSON instructs how to create Value fields from raw bytes
// TODO(nsiow) break up this function
func (v *Value) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("value too short: %s", data)
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return fmt.Errorf("value too short: %s", data)
	}

	value, ok, err := parseScalarValue(data)
	if err != nil {
		return err
	}
	if ok {
		if value == "" && strings.TrimSpace(string(data)) == "null" {
			*v = []string{}
		} else {
			*v = []string{value}
		}
		return nil
	}

	switch {
	// Handle multi-value case
	case data[0] == '[':
		var raw []stdjson.RawMessage
		err := stdjson.Unmarshal(data, &raw)
		if err != nil {
			return fmt.Errorf("error in multi-value clause of Value:\nerror=%s\ndata=%v", err, data)
		}
		a := make([]string, len(raw))
		for i, item := range raw {
			value, ok, err := parseScalarValue(item)
			if err != nil {
				return err
			}
			if !ok || strings.TrimSpace(string(item)) == "null" {
				return fmt.Errorf("array values should be string, number, or bool:\ndata=%s", item)
			}
			a[i] = value
		}
		*v = a
		return nil
	// Anything else is an error
	default:
		return fmt.Errorf("should be string, number, bool, or []value, but received invalid input:\ndata=%s", data)
	}
}

func parseScalarValue(data []byte) (string, bool, error) {
	trimmed := strings.TrimSpace(string(data))

	// Check for null case
	if trimmed == "null" {
		return "", true, nil
	}

	// Check for true/false
	if trimmed == "true" {
		return "true", true, nil
	}
	if trimmed == "false" {
		return "false", true, nil
	}

	// Handle single-value case
	if strings.HasPrefix(trimmed, `"`) {
		var s string
		err := json.Unmarshal([]byte(trimmed), &s)
		if err != nil {
			return "", false, fmt.Errorf("error in single-value clause of Value:\nerror=%s\ndata=%v", err, data)
		}
		return s, true, nil
	}

	if stdjson.Valid([]byte(trimmed)) && len(trimmed) > 0 &&
		((trimmed[0] >= '0' && trimmed[0] <= '9') || trimmed[0] == '-') {
		return trimmed, true, nil
	}

	return "", false, nil
}

// Count returns the number of strings represented in the Value
func (v *Value) Count() int {
	return len(*v)
}

// Empty returns whether or not the Value contains any values
func (v *Value) Empty() bool {
	return v.Count() == 0
}

// Contains returns whether or not the provided string is a member of the Value
func (v *Value) Contains(s string) bool {
	return slices.Contains(*v, s)
}
