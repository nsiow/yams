package smartrw

import (
	"bytes"
	"testing"
)

func TestStdoutWriter(t *testing.T) {
	var buf []byte

	// create a new test writer
	writer := NewStdoutWriter()

	// redirect stdout
	writer.stdout = bytes.NewBuffer(buf)

	// test writes
	n, err := writer.Write([]byte("foo"))
	if err != nil {
		t.Fatalf("unexpected failure when writing: %v", err)
	}
	if n != 3 {
		t.Fatalf("expected to write 3 bytes but observed: %d", n)
	}

	n, err = writer.Write([]byte("bar baz"))
	if err != nil {
		t.Fatalf("unexpected failure when writing: %v", err)
	}
	if n != 7 {
		t.Fatalf("expected to write 7 bytes but observed: %d", n)
	}

	// test close
	err = writer.Close()
	if err != nil {
		t.Fatalf("unexpected failure when closing: %v", err)
	}
}
