package sim

import (
	"slices"
	"testing"
)

// TestTrace does basic functionality tests of the Trace objects
func TestTrace(t *testing.T) {
	trc := Trace{}
	if len(trc.Messages()) != 0 {
		t.Fatal("should have been 0 messages in new trace")
	}

	s := "testing testing 123"
	trc.Add(s)
	if len(trc.Messages()) != 1 {
		t.Fatal("should have been 1 message in trace")
	}
	if !slices.Contains(trc.Messages(), s) {
		t.Fatalf("trace should contain %s", s)
	}

	s = "another message"
	trc.Add(s)
	if len(trc.Messages()) != 2 {
		t.Fatal("should have been 2 messages in trace")
	}
	if !slices.Contains(trc.Messages(), s) {
		t.Fatalf("trace should contain %s", s)
	}
}
