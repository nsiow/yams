package sim

// SimulationOptions contains all possible customizatons for simulator logic + runtime
type SimOptions struct {
	// FailOnUnknownCondition determines whether or not to fail on unknown Condition evaluation
	FailOnUnknownCondition bool
}

// Option implements the functional options pattern for simulator options
type Option func(*SimOptions) error

// WithFailOnUnknownCondition causes simulation to fail if we encounter a Conditon we do not know
// how to handle
func WithFailOnUnknownCondition() Option {
	return func(opt *SimOptions) error {
		opt.FailOnUnknownCondition = true
		return nil
	}
}
