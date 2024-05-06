package policy

import (
	"encoding/json"
	"fmt"
)

// Value is a JSON-centric helper struct to facilitate one-or-more value representations
type Value []string

// NewValue creates a new PolicyString struct using the supplied values
func NewValue(values ...string) Value {
	return values
}

// UnmarshalJSON instructs how to create Value fields from raw bytes
func (v *Value) UnmarshalJSON(data []byte) error {
	// We should have either a string (""), an array ([]), or null (null); anything shorter is invalid
	if len(data) < 2 {
		return fmt.Errorf("value too short: %s", string(data))
	}

	// Check for null case
	if len(data) == 4 && string(data) == "null" {
		*v = []string{}
		return nil
	}

	switch {
	// Handle single-value case
	case data[0] == '"':
		var s string
		err := json.Unmarshal(data, &s)
		// TODO(nsiow) figure out correct behavior of empty string; IAM treates it... weirdly
		if err != nil || len(s) == 0 {
			return fmt.Errorf("error in single-value clause of Value:\ndata=%s\nerror=%v", string(data), err)
		}
		a := []string{s}
		*v = a
		return nil
	// Handle multi-value case
	case data[0] == '[':
		var a []string
		err := json.Unmarshal(data, &a)
		if err != nil {
			return fmt.Errorf("error in multi-value clause of Value:\ndata=%s\nerror=%v", string(data), err)
		}
		*v = a
		return nil
	// Anything else is an error
	default:
		return fmt.Errorf("should be string or []string, but received invalid input:\ndata=%s", string(data))
	}
}

// Count returns the number of strings represented in the Value
func (v *Value) Count() int {
	return len(*v)
}

// Empty returns whether or not the Value contains any values
func (v *Value) Empty() bool {
	return v.Count() == 0
}
