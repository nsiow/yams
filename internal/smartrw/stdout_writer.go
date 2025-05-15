package smartrw

import (
	"io"
	"os"
)

// StdoutWriter is a basic io.WriteCloser that writes to stdout and ignores Close() calls
type StdoutWriter struct {
	stdout io.Writer
}

// NewStdoutWriter creates and returns a new writer appropriate for writing to stdout
func NewStdoutWriter() *StdoutWriter {
	return &StdoutWriter{stdout: os.Stdout}
}

func (s *StdoutWriter) Write(p []byte) (n int, err error) {
	return s.stdout.Write(p)
}

func (s *StdoutWriter) Close() error {
	return nil
}
