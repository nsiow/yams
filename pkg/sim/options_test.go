package sim

import (
	"testing"
)

func TestWithSkipUnknownCondition(t *testing.T) {
	// Apply option
	opt := Options{}
	f := WithSkipUnknownConditionOperators()
	err := f(&opt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check results
	if opt.SkipUnknownConditionOperators != true {
		t.Fatalf("expected: %v, got: %v", true, opt.SkipUnknownConditionOperators)
	}
}
