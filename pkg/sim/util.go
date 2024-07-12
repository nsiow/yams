package sim

import "strconv"

// Id is a helper function which takes in an index and an ID, returning the ID if it's non-empty
// and the index otherwise
//
// It's most commonly used to resolve a valid identifier for a statement or policy, where a Policy
// ID or Sid is preferable but a relative index is a valid fallback
func Id(id string, idx int) string {
	if len(id) > 0 {
		return id
	}

	return strconv.Itoa(idx)
}
