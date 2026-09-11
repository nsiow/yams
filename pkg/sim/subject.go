package sim

import (
	"github.com/nsiow/yams/pkg/aws/sar/types"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/trace"
)

// subject is a type representing the smallest simulatable structure.
//
// It contains both the data required for auth simulation as well as any accessory data
type subject struct {
	auth  AuthContext
	opts  Options
	trc   trace.Trace
	extra Extra

	policyVersion string

	cachedAction  *types.Action
	actionMatches map[*policy.Statement]bool
}

// newSubject creates a new `subject` struct with the provided authorization context and options
func newSubject(ac AuthContext, opts Options) subject {
	subj := subject{
		auth:          ac,
		opts:          opts,
		actionMatches: make(map[*policy.Statement]bool),
	}

	if opts.EnableTracing {
		subj.trc.Enable()
	}

	return subj
}
