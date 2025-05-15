package sim

import "github.com/nsiow/yams/pkg/sim/trace"

// SimResult defines the output of a policy simulation option
type SimResult struct {
	// Principal corresponds to the ARN of the Principal used for this evaluation
	Principal string

	// Action corresponds to the AWS API action used for this evaluation
	Action string

	// Resource corresponds to the ARN of the Resource used for this evaluation
	Resource string

	// IsAllowed corresponds to whether or not the operation was allowed
	IsAllowed bool

	// Trace contains an evaluation trace providing context as to the access evaluation process
	Trace *trace.Trace
}
