package smartrw

import (
	"compress/gzip"
	"io"
)

type GzipReadCloser struct {
	r  io.ReadCloser
	gz *gzip.Reader
}

func NewGzipReadCloser(r io.ReadCloser) (*GzipReadCloser, error) {
	wrapped, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	gz := GzipReadCloser{
		r:  r,
		gz: wrapped,
	}
	return &gz, nil
}

func (g *GzipReadCloser) Read(p []byte) (n int, err error) {
	return g.gz.Read(p)
}

func (g *GzipReadCloser) Close() error {
	err := g.gz.Close()
	if err != nil {
		return err
	}

	return g.r.Close()
}
