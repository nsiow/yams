package sim

import (
	"reflect"
	"testing"

	"github.com/nsiow/yams/pkg/entities"
)

func TestWithEnvironment(t *testing.T) {
	// Define environment
	opt := options{}
	env := entities.Environment{Principals: []entities.Principal{
		{
			Arn: "foo",
		},
	}}

	// Apply option
	f := WithEnvironment(&env)
	err := f(&opt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check results
	if !reflect.DeepEqual(env, *opt.Environment) {
		t.Fatalf("expected: %v, got: %v", env, opt.Environment)
	}
}

func TestWithFailOnUnknownCondition(t *testing.T) {
	// Apply option
	opt := options{}
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
