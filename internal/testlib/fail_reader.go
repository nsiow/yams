package testlib

import "fmt"

// FailReader conforms to [io.Reader] and always fails on Read(...) for use with testing
type FailReader struct{}

func (f *FailReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("FailReader: dutifully failing")
}
