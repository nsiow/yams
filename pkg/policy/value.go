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
		return fmt.Errorf("value too short: %s", data)
	}

	// Check for null case
	if len(data) == 4 && string(data) == "null" {
		*v = []string{}
		return nil
	}

	// Check for true/false
	if len(data) == 4 && string(data) == "true" {
		*v = []string{"true"}
		return nil
	}
	if len(data) == 5 && string(data) == "false" {
		*v = []string{"false"}
		return nil
	}

	switch {
	// Handle single-value case
	case data[0] == '"':
		var s string
		err := json.Unmarshal(data, &s)
		// TODO(nsiow) this accounts for some malformed policies, but perhaps is addressable elsewhere?
		if err != nil || s == `"` {
			return fmt.Errorf("error in single-value clause of Value:\nerror=%s\ndata=%v", err, data)
		}
		a := []string{s}
		*v = a
		return nil
	// Handle multi-value case
	case data[0] == '[':
		var a []string
		err := json.Unmarshal(data, &a)
		if err != nil {
			return fmt.Errorf("error in multi-value clause of Value:\nerror=%s\ndata=%v", err, data)
		}
		*v = a
		return nil
	// Anything else is an error
	default:
		return fmt.Errorf("should be string or []string, but received invalid input:\ndata=%s", data)
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
