package sim

// Result defines the output of a policy simulation option
type Result struct {
	// IsAllowed corresponds to whether or not the operation was allowed
	IsAllowed bool

	// ResultContext
	ResultContext ResultContext
}
