package smartrw

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

type Writer struct {
	Dest string
	io.WriteCloser
}

func NewWriter(dest string) (io.WriteCloser, error) {
	w, err := selectWriter(dest)
	if err != nil {
		return nil, err
	}

	return &Writer{
		Dest:        dest,
		WriteCloser: w,
	}, nil
}

func selectWriter(dest string) (io.WriteCloser, error) {
	// empty = write to stdout
	if len(dest) == 0 {
		return NewStdoutWriter(), nil
	}

	// handle file:// and protocol-less destinations
	if strings.HasPrefix(dest, "file://") || !strings.Contains(dest, "://") {
		dest = strings.TrimPrefix(dest, "file://")
		w, err := os.Create(dest)
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(dest, ".gz") {
			return gzip.NewWriter(w), nil
		}

		return w, nil
	}

	// handle unknown protocols
	idx := strings.Index(dest, "://")
	protocol := dest[:idx]
	return nil, fmt.Errorf("unknown smartrw protocol: %s", protocol)
}
