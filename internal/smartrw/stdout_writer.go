package smartrw

import "os"

// StdoutWriter is a basic io.WriteCloser that writes to stdout and ignores Close() calls
type StdoutWriter struct{}

func (s *StdoutWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (s *StdoutWriter) Close() error {
	return nil
}
