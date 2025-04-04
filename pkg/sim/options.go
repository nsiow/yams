package sim

// Options contains all possible customizatons for simulator logic + behavior
type Options struct {
	// SkipUnknownConditionOperators causes simulation to skip over unknown condition operators
	SkipUnknownConditionOperators bool

	// SkipUnknownConditionKeys causes simulation to skip over unknown condition keys
	SkipUnknownConditionKeys bool
}

// OptionF implements the functional options pattern for simulator options
type OptionF func(*Options) error

// WithSkipUnknownConditionOperators toggles SkipUnknownConditionOperators to true
func WithSkipUnknownConditionOperators() OptionF {
	return func(opt *Options) error {
		opt.SkipUnknownConditionOperators = true
		return nil
	}
}

// WithSkipUnknownConditionKeys toggles SkipUnknownConditionKeys to true
func WithSkipUnknownConditionKeys() OptionF {
	return func(opt *Options) error {
		opt.SkipUnknownConditionKeys = true
		return nil
	}
}
