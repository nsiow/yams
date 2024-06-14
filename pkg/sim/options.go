package sim

// Sim contains all possible customizatons for simulator logic + runtime
type Options struct {
	// FailOnUnknownCondition determines whether or not to fail on unknown Condition evaluation
	// TODO(nsiow) make opts.FailOnUnknownCondition default
	FailOnUnknownCondition bool
}

// OptionF implements the functional options pattern for simulator options
type OptionF func(*Options) error

// WithFailOnUnknownCondition causes simulation to fail if we encounter a Conditon we do not know
// how to handle
func WithFailOnUnknownCondition() OptionF {
	return func(opt *Options) error {
		opt.FailOnUnknownCondition = true
		return nil
	}
}
