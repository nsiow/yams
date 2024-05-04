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
	// First make sure the data can be marshalled at all
	var raw any
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return fmt.Errorf("unable to parse:\nvalue = %s\nerror = %v", string(data), err)
	}

	// Handle the different cases between both types
	value := []string{}
	switch cast := raw.(type) {
	case string:
		value = []string{cast}
	case []any:
		for _, a := range cast {
			s, ok := a.(string)
			if !ok {
				return fmt.Errorf("should be string or []string, saw %T for %v", a, a)
			}
			value = append(value, s)
		}
	case nil:
		break
	default:
		return fmt.Errorf("should be string or []string, saw %T for %v", cast, cast)
	}

	*v = value
	return nil
}

// Count returns the number of strings represented in the Value
func (v *Value) Count() int {
	return len(*v)
}

// Empty returns whether or not the Value contains any values
func (v *Value) Empty() bool {
	return v.Count() == 0
}
