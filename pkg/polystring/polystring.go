package polystring

import "encoding/json"

// PolyString is a JSON-centric helper struct to facilitate one-or-more value representations
type PolyString struct {
	Values []string
}

// NewPolyString creates a new PolicyString struct using the supplied values
func NewPolyString(values ...string) PolyString {
	return PolyString{Values: values}
}

// UnmarshalJSON instructs how to create PolyString fields from raw bytes
func (p *PolyString) UnmarshalJSON(data []byte) error {
	// Handle empty string
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	// If it looks like an array; handle it as such
	if data[0] == '[' && data[len(data)-1] == ']' {
		return json.Unmarshal(data, &p.Values)
	}

	// Otherwise handle it as a string
	p.Values = append(p.Values, string(data))
	return nil
}

// Count returns the number of strings represented in the PolyString
func (p *PolyString) Count() int {
	return len(p.Values)
}

// Empty returns whether or not the PolyString contains any values
func (p *PolyString) Empty() bool {
	return p.Count() == 0
}
