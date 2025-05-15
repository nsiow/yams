package smartrw

import (
	"compress/gzip"
	"io"
)

type GzipWriteCloser struct {
	r  io.WriteCloser
	gz *gzip.Writer
}

func NewGzipWriteCloser(r io.WriteCloser) *GzipWriteCloser {
	wrapped := gzip.NewWriter(r)

	gz := GzipWriteCloser{
		r:  r,
		gz: wrapped,
	}
	return &gz
}

func (g *GzipWriteCloser) Write(p []byte) (n int, err error) {
	return g.gz.Write(p)
}

func (g *GzipWriteCloser) Close() error {
	err := g.gz.Close()
	if err != nil {
		return err
	}

	return g.r.Close()
}
