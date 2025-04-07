package sim

import "github.com/nsiow/yams/pkg/sim/trace"

// subject is a type representing the smallest simulatable structure.
//
// It contains both the data required for auth simulation as well as any accessory data and
// structured
type subject struct {
	ac   *AuthContext
	opts *Options
	trc  *trace.Trace
}

// newSubject creates a new `subject` struct with the provided authorization context and options
func newSubject(ac *AuthContext, opts *Options) *subject {
	return &subject{
		ac:   ac,
		opts: opts,
		trc:  trace.New(),
	}
}
