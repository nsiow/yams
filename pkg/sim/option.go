package sim

// options contains all possible customizatons for simulator logic + runtime
type options struct {
	// FailOnUnknownCondition determines whether or not to fail on unknown Condition evaluation
	FailOnUnknownCondition bool
}

// Option implements the functional options pattern for simulator options
type Option func(*options) error

// WithFailOnUnknownCondition causes simulation to fail if we encounter a Conditon we do not know
// how to handle
func WithFailOnUnknownCondition(o *options) Option {
	return func(opt *options) error {
		opt.FailOnUnknownCondition = true
		return nil
	}
}
