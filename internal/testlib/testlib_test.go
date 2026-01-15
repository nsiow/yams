package testlib

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// mockT is a mock testing.T for testing assertion failures
type mockT struct {
	fatalCalled bool
	logCalled   bool
	messages    []string
}

func (m *mockT) Helper()                         {}
func (m *mockT) Run(_ string, f func(*testing.T)) bool { return true }
func (m *mockT) Fatalf(format string, args ...any) {
	m.fatalCalled = true
}
func (m *mockT) Logf(format string, args ...any) {
	m.logCalled = true
}

func TestTestTime(t *testing.T) {
	result := TestTime()

	// Verify it returns a consistent time
	expected, _ := time.Parse(time.DateTime, time.DateTime)
	if !result.Equal(expected) {
		t.Fatalf("TestTime returned %v, expected %v", result, expected)
	}

	// Verify calling it multiple times gives same result
	result2 := TestTime()
	if !result.Equal(result2) {
		t.Fatal("TestTime should return consistent values")
	}
}

func TestRunTestSuite_Success(t *testing.T) {
	tests := []TestCase[int, int]{
		{Name: "add_one", Input: 1, Want: 2},
		{Name: "add_two", Input: 2, Want: 3},
		{Input: 5, Want: 6}, // no name - tests default name generation
	}

	RunTestSuite(t, tests, func(i int) (int, error) {
		return i + 1, nil
	})
}

func TestRunTestSuite_ExpectedError(t *testing.T) {
	tests := []TestCase[int, int]{
		{Name: "should_error", Input: -1, ShouldErr: true},
	}

	RunTestSuite(t, tests, func(i int) (int, error) {
		if i < 0 {
			return 0, errors.New("negative number")
		}
		return i, nil
	})
}

func TestGenerateFailureOutput(t *testing.T) {
	tc := TestCase[string, string]{
		Name:  "test_case",
		Input: "input_value",
		Want:  "expected_value",
	}

	output := generateFailureOutput(tc, "actual_value")

	// Verify output contains expected elements
	if !strings.Contains(output, "test_case") {
		t.Error("output should contain test case name")
	}
	if !strings.Contains(output, "input_value") {
		t.Error("output should contain input value")
	}
	if !strings.Contains(output, "expected_value") {
		t.Error("output should contain expected value")
	}
	if !strings.Contains(output, "actual_value") {
		t.Error("output should contain actual value")
	}
	if !strings.Contains(output, "delta") {
		t.Error("output should contain diff command")
	}
}

func TestGenerateFailureOutput_NoName(t *testing.T) {
	tc := TestCase[string, string]{
		Name:  "",
		Input: "input",
		Want:  "want",
	}

	output := generateFailureOutput(tc, "got")

	// Verify default name handling
	if !strings.Contains(output, "noname") {
		t.Error("output should contain default name 'noname'")
	}
}

func TestPrettyPrint(t *testing.T) {
	type testStruct struct {
		A string
		B int
	}

	input := testStruct{A: "hello", B: 42}
	output := prettyPrint(input)

	// Verify formatting is applied
	if !strings.Contains(output, "{\n\t") {
		t.Error("output should contain formatted braces")
	}
	if !strings.Contains(output, ",\n") {
		t.Error("output should have newlines after commas")
	}
}

func TestAssertPanic(t *testing.T) {
	t.Run("catches_panic", func(t *testing.T) {
		defer AssertPanic(t)
		panic("test panic")
	})
}

func TestAssertPanicWithText(t *testing.T) {
	t.Run("catches_matching_panic", func(t *testing.T) {
		defer AssertPanicWithText(t, "expected.*panic")
		panic("expected test panic")
	})
}

func TestFailReader_Read(t *testing.T) {
	fr := &FailReader{}
	buf := make([]byte, 10)

	n, err := fr.Read(buf)

	if n != 0 {
		t.Errorf("expected 0 bytes read, got %d", n)
	}
	if err == nil {
		t.Error("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "FailReader") {
		t.Error("error message should contain 'FailReader'")
	}
}
