package smartrw

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Reader struct {
	Source string
	io.ReadCloser
}

func NewReader(src string) (*Reader, error) {
	reader := Reader{Source: src}
	err := reader.Reset()
	return &reader, err
}

func (r *Reader) Reset() error {
	rc, err := selectReader(r.Source)
	if err != nil {
		return err
	}

	r.ReadCloser = rc
	return nil
}

func selectReader(src string) (io.ReadCloser, error) {
	var r io.ReadCloser

	// Determine base reader and configuration
	switch {

	// handle file:// and protocol-less sources
	case strings.HasPrefix(src, "file://") || !strings.Contains(src, "://"):
		f, err := os.Open(strings.TrimPrefix(src, "file://"))
		if err != nil {
			return nil, fmt.Errorf("unable to open file: %v", err)
		}
		r = f

	// handle s3:// sources
	case strings.HasPrefix(src, "s3://"):
		s3, err := NewS3Reader(strings.TrimPrefix(src, "s3://"))
		if err != nil {
			return nil, fmt.Errorf("unable to create s3 reader: %v", err)
		}
		r = s3

	// handle unknown protocols
	default:
		return nil, fmt.Errorf("unknown smartrw protocol: %s", src)
	}

	// Wrappers for compression, etc
	if strings.HasSuffix(src, ".gz") {
		gz, err := NewGzipReadCloser(r)
		if err != nil {
			r.Close()
			return nil, fmt.Errorf("unable to wrap gzip reader: %v", err)
		}
		r = gz
	}

	return r, nil
}
