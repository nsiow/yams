package testlib

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestTime is a helper function that allows us to use a specific time across all tests
func TestTime() time.Time {
	t, err := time.Parse(time.DateTime, time.DateTime)
	if err != nil {
		panic("somehow failed to generate a reference time for testing")
	}

	return t
}

// TestCase defines a single executable test case meant to be part of a suite
type TestCase[I, O any] struct {
	Name  string
	Input I
	Want  O
	// TODO(nsiow) switch this to an error type and use errors.Is(...)
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
			t.Helper()
			got, err := f(tc.Input)

			switch {
			case err == nil && tc.ShouldErr:
				t.Fatalf("expected error, got success")
			case err != nil && tc.ShouldErr:
				// expected error; got error
				t.Logf("test saw expected error: %v", err)
				return
			case err == nil && !tc.ShouldErr:
				if !reflect.DeepEqual(tc.Want, got) {
					msg := GenerateFailureOutput(tc, got)
					t.Fatal(msg)
				}
			case err != nil && !tc.ShouldErr:
				t.Fatalf("unexpected error during test case: %v", err)
			default:
				panic("should never reach this condition")
			}
		})
	}
}

// GenerateFailureOutput creates a more usable/readable "wanted vs got" diff for tests
func GenerateFailureOutput[I, O any](tc TestCase[I, O], got any) string {
	header := "// --------------------------------------------------"
	tmpdir := os.TempDir()

	wantedJson, _ := json.MarshalIndent(tc.Want, "", "  ")
	gotJson, _ := json.MarshalIndent(got, "", "  ")

	wantedMessage := "unable to generate for output for `wanted`"
	wantedFile := path.Join(tmpdir, fmt.Sprintf("yams.%s.wanted.json", tc.Name))
	err := os.WriteFile(wantedFile, wantedJson, 0644)
	if err == nil {
		wantedMessage = fmt.Sprintf("expected output available @ %s", wantedFile)
	}

	gotMessage := "unable to generate for output for `got`"
	gotFile := path.Join(tmpdir, fmt.Sprintf("yams.%s.got.json", tc.Name))
	err = os.WriteFile(gotFile, gotJson, 0644)
	if err == nil {
		gotMessage = fmt.Sprintf("observed output available @ %s", gotFile)
	}

	return strings.Join([]string{
		"test case failed",
		strings.Join([]string{header, "// input", header}, "\n"),
		fmt.Sprintf("%#v", tc.Input),
		strings.Join([]string{header, "// wanted", header}, "\n"),
		fmt.Sprintf("%#v", tc.Want),
		strings.Join([]string{header, "// got", header}, "\n"),
		fmt.Sprintf("%#v", got),
		strings.Join([]string{header, "// wanted (pretty)", header}, "\n"),
		fmt.Sprintf(wantedMessage),
		strings.Join([]string{header, "// got (pretty)", header}, "\n"),
		fmt.Sprintf(gotMessage),
	}, "\n")
}
