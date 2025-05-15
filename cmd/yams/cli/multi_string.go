package cli

import "strings"

// MultiString implements the flag.Value interface for multiple string values.
type MultiString []string

func (m *MultiString) String() string {
	return strings.Join(*m, ", ")
}

func (m *MultiString) Set(value string) error {
	*m = append(*m, value)
	return nil
}
