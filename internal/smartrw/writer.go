package smartrw

import (
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

	var w io.WriteCloser

	// Determine base reader and configuration
	switch {

	// handle file:// and protocol-less sources
	case strings.HasPrefix(dest, "file://") || !strings.Contains(dest, "://"):
		f, err := os.Create(strings.TrimPrefix(dest, "file://"))
		if err != nil {
			return nil, fmt.Errorf("unable to open file: %v", err)
		}
		w = f

	// handle s3:// sources
	case strings.HasPrefix(dest, "s3://"):
		s3, err := NewS3Writer(strings.TrimPrefix(dest, "s3://"))
		if err != nil {
			return nil, fmt.Errorf("unable to create s3 reader: %v", err)
		}
		w = s3

	// handle unknown protocols
	default:
		return nil, fmt.Errorf("unknown smartrw protocol: %s", dest)
	}

	// Wrappers for compression, etc
	if strings.HasSuffix(dest, ".gz") {
		w = NewGzipWriteCloser(w)
	}

	return w, nil
}
