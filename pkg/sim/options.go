package sim

import "github.com/nsiow/yams/pkg/entities"

// DEFAULT_OPTIONS uses all default configuration options for simulation
var DEFAULT_OPTIONS = NewOptions()

// Options contains all possible customizatons for simulator logic + behavior
type Options struct {
	// SkipServiceAuthorizationValidation foregoes the usual validation via the Service Authorization
	// Reference. This will result in faster simulation but at the cost of real-world accuracy
	SkipServiceAuthorizationValidation bool

	// EnableTracing turns on active tracing for requests. This incurs a minor performance penalty but
	// allows for helpful explanations of how a particular simulation result was achieved
	EnableTracing bool

	// Context specifies additional key/value pairs that should be carried along in the Authorization
	// context
	Context Bag[string]

	// Overlays allows one to specify a special "overlay" Universe in which entity lookup takes place
	// over the primary simulation Universe
	Overlay *entities.Universe

	// DefaultS3Key specifies which S3 object key should be used to expand S3 bucket ARNs by default.
	// In other words, it enables simulation against S3 object-level calls for operations where
	// individual object keys cannot be provided
	DefaultS3Key string

	// EnableFuzzyMatchArn enables fuzzy-matching for principal/resource values. This will do a
	// case-insensitive search based on user inputs, and return an error if more than one value
	// matches
	EnableFuzzyMatchArn bool

	// Strict causes simulation to fail if it encounters an AWS managed policy or group that it cannot
	// resolve. Otherwise, simulation will use an empty policy
	Strict bool

	// ForceFailure causes all simulation results to throw an error. Primarily used for testing and
	// debugging purposes
	ForceFailure bool
}

// NewOptions creates and returns a new Options struct parameterized with the provided options
func NewOptions(funcs ...OptionF) Options {
	o := Options{
		DefaultS3Key: "*",
	}

	for _, f := range funcs {
		f(&o)
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

// WithTracing toggles EnableTracing to true
func WithTracing() OptionF {
	return func(opt *Options) {
		opt.EnableTracing = true
	}
}

// WithOverlay adds the provided "overlay" universe to our options
func WithOverlay(overlays *entities.Universe) OptionF {
	return func(opt *Options) {
		opt.Overlay = overlays
	}
}

// WithAdditionalProperties adds the provided properties to the request context
func WithAdditionalProperties(props map[string]string) OptionF {
	return func(opt *Options) {
		opt.Context = NewBagFromMap(props)
	}
}

// WithDefaultS3Key sets the provided S3 key as the default for all buckets
func WithDefaultS3Key(key string) OptionF {
	return func(opt *Options) {
		opt.DefaultS3Key = key
	}
}

// WithEnableFuzzyMatchArn turns on fuzzy-matching for ARN values
func WithEnableFuzzyMatchArn() OptionF {
	return func(opt *Options) {
		opt.EnableFuzzyMatchArn = true
	}
}

// WithStrict enables failures for unknown managed policies and groups
func WithStrict() OptionF {
	return func(opt *Options) {
		opt.Strict = true
	}
}

// WithForceFailure causes all simulations to fail
func WithForceFailure() OptionF {
	return func(opt *Options) {
		opt.ForceFailure = true
	}
}

// TestingSimulationOptions provides a specific set of simulation options appropriate for most
// tests. It allows for exercising difficult-to-reach error paths while also allowing us to bend
// the rules a bit for testing -- fewer checks around the specifics of the dummy resource calls we
// use
var TestingSimulationOptions = NewOptions(
	WithStrict(),
	WithTracing(),
	WithSkipServiceAuthorizationValidation(),
)
