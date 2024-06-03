package sim

import "github.com/nsiow/yams/pkg/entities"

// options contains all possible customizatons for simulator logic + runtime
type options struct {
	// Environment defines the world known to the simulator
	Environment *entities.Environment

	// FailOnUnknownCondition determines whether or not to fail on unknown Condition evaluation
	FailOnUnknownCondition bool
}

// Option implements the functional options pattern for simulator options
type Option func(*options) error

// WithEnvironment has the Simulator use the provided Environment on startup
func WithEnvironment(env *entities.Environment) Option {
	return func(opt *options) error {
		opt.Environment = env
		return nil
	}
}

// WithFailOnUnknownCondition causes simulation to fail if we encounter a Conditon we do not know
// how to handle
func WithFailOnUnknownCondition() Option {
	return func(opt *options) error {
		opt.FailOnUnknownCondition = true
		return nil
	}
}
