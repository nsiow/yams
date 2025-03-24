package testlib

import (
	"regexp"
	"testing"
)

// AssertPanic asserts that the current function panics (with any panic text)
//
// It is meant to be used within a defer statement within a test function
func AssertPanic(t *testing.T) {
	AssertPanicWithText(t, ".*")
}

// AssertPanicWithText asserts that the current function panics and matches the provided text
//
// It is meant to be used within a defer statement within a test function
func AssertPanicWithText(t *testing.T, pattern string) {
	if r := recover(); r == nil {
		t.Fatalf("expected panic but observed success")
	} else {
		panicstr := r.(string)
		match, err := regexp.MatchString(pattern, panicstr)
		if err != nil {
			t.Fatalf("unable to apply panic rgx: %s", pattern)
		}

		if match {
			t.Logf("saw expected panic: %s", panicstr)
		} else {
			t.Fatalf("observed different panic (%s) than expected: %s", panicstr, pattern)
		}
	}
}
