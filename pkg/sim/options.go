package sim

// Options contains all possible customizatons for simulator logic + behavior
type Options struct {
	// SkipServiceAuthorizationValidation foregoes the usual validation via the Service Authoization
	// Reference. This will result in faster simulation but at the cost of real-world accuracy
	SkipServiceAuthorizationValidation bool
}

// NewOptions creates and returns a new Options struct parameterized with the provided options
func NewOptions(funcs ...OptionF) *Options {
	o := &Options{}
	for _, f := range funcs {
		f(o)
	}
	return o
}

// OptionF implements the functional options pattern for simulator options
type OptionF func(*Options)

// WithSkipServiceAuthorizationValidation toggles SkipServiceAuthorizationValidation to true
func WithSkipServiceAuthorizationValidation() OptionF {
	return func(opt *Options) {
		opt.SkipServiceAuthorizationValidation = true
	}
}

// TestingSimulationOptions provides a specific set of simulation options appropriate for most
// tests. It allows for exercising difficult-to-reach error paths while also allowing us to bend
// the rules a bit for testing -- fewer checks around the specifics of the dummy resource calls we
// use
var TestingSimulationOptions = NewOptions(
	WithSkipServiceAuthorizationValidation(),
)
