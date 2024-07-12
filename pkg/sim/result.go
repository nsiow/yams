package sim

import "github.com/nsiow/yams/pkg/sim/trace"

// Result defines the output of a policy simulation option
type Result struct {
	// IsAllowed corresponds to whether or not the operation was allowed
	IsAllowed bool

	// Trace contains an evaluation trace providing context as to the access evaluation process
	Trace *trace.Trace
}
