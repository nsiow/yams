package sim

import (
	"testing"
)

func TestWithFailOnUnknownCondition(t *testing.T) {
	// Apply option
	opt := SimOptions{}
	f := WithFailOnUnknownCondition()
	err := f(&opt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check results
	if opt.FailOnUnknownCondition != true {
		t.Fatalf("expected: %v, got: %v", true, opt.FailOnUnknownCondition)
	}
}
