package testrunner

import (
	"reflect"
	"testing"
)

// TestCase defines a single executable test case meant to be part of a suite
type TestCase[I, O any] struct {
	Name      string
	Input     I
	Want      O
	ShouldErr bool
}

// RunTestSuite executes a provided table of tests
func RunTestSuite[I, O any](
	t *testing.T,
	testCases []TestCase[I, O],
	f func(I) (O, error)) {

	t.Helper()
	for _, tc := range testCases {
		tc := tc // local variable in case we need to use pointer to loop var
		t.Run(tc.Name, func(t *testing.T) {
			got, err := f(tc.Input)

			switch {
			case err == nil && tc.ShouldErr:
				t.Fatalf("expected error, got success")
			case err != nil && tc.ShouldErr:
				// expected error; got error
				t.Logf("test saw expected error: %v", err)
				return
			case err == nil && !tc.ShouldErr:
				if !reflect.DeepEqual(got, tc.Want) {
					t.Fatalf("failed test case; wanted %+v got %+v", tc.Want, got)
				}
			case err != nil && !tc.ShouldErr:
				t.Fatalf("unexpected error during test case: %v", err)
			default:
				panic("should never reach this condition")
			}
		})
	}
}
