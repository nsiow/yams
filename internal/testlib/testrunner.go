package testlib

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
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
// MAYBE(nsiow) wrap in a constructor that requires names for tests?
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
	for i, tc := range testCases {

		// If a name isn't provided, use the index instead
		if len(tc.Name) == 0 {
			tc.Name = strconv.Itoa(i)
		}

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
				// IDEA(nsiow) make the comparison function configurable
				if !reflect.DeepEqual(tc.Want, got) {
					msg := generateFailureOutput(tc, got)
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

// generateFailureOutput creates a more usable/readable "wanted vs got" diff for tests
func generateFailureOutput[I, O any](tc TestCase[I, O], got any) string {
	header := "--------------------------------------------------"
	tmpdir := os.TempDir()

	prettyWanted := prettyPrint(tc.Want)
	prettyGot := prettyPrint(got)

	if len(tc.Name) == 0 {
		tc.Name = "noname"
	}

	wantedMessage := "unable to generate for output for `wanted`"
	wantedFile := path.Join(tmpdir, fmt.Sprintf("yams.%s.wanted.debug", tc.Name))
	err := os.WriteFile(wantedFile, []byte(prettyWanted), 0644)
	if err == nil {
		wantedMessage = fmt.Sprintf("expected output available @ %s", wantedFile)
	}

	gotMessage := "unable to generate for output for `got`"
	gotFile := path.Join(tmpdir, fmt.Sprintf("yams.%s.got.json", tc.Name))
	err = os.WriteFile(gotFile, []byte(prettyGot), 0644)
	if err == nil {
		gotMessage = fmt.Sprintf("observed output available @ %s", gotFile)
	}

	return strings.Join([]string{
		"test case failed",
		strings.Join([]string{header, "|\tname", header}, "\n"),
		tc.Name,
		strings.Join([]string{header, "|\tinput", header}, "\n"),
		prettyPrint(tc.Input),
		strings.Join([]string{header, "|\twanted", header}, "\n"),
		prettyPrint(tc.Want),
		strings.Join([]string{header, "|\tgot", header}, "\n"),
		prettyPrint(got),
		strings.Join([]string{header, "|\twanted (pretty)", header}, "\n"),
		fmt.Sprint(wantedMessage),
		strings.Join([]string{header, "|\tgot (pretty)", header}, "\n"),
		fmt.Sprint(gotMessage),
		strings.Join([]string{header, "|\tdiff command", header}, "\n"),
		fmt.Sprintf("delta %s %s", gotFile, wantedFile),
	}, "\n")
}

func prettyPrint(obj any) string {
	pretty := fmt.Sprintf("%#v", obj)
	pretty = strings.ReplaceAll(pretty, ",", ",\n")
	pretty = strings.ReplaceAll(pretty, "{", "{\n\t")
	pretty = strings.ReplaceAll(pretty, "}", "\t\n}")

	return pretty
}
