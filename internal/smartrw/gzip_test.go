package smartrw

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

// mockReadCloser allows testing of GzipReadCloser
type mockReadCloser struct {
	data      *bytes.Buffer
	closed    bool
	closeErr  error
	readCount int
}

func newMockReadCloser(data []byte) *mockReadCloser {
	return &mockReadCloser{data: bytes.NewBuffer(data)}
}

func (m *mockReadCloser) Read(p []byte) (int, error) {
	m.readCount++
	return m.data.Read(p)
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	return m.closeErr
}

// mockWriteCloser allows testing of GzipWriteCloser
type mockWriteCloser struct {
	data       bytes.Buffer
	closed     bool
	closeErr   error
	writeCount int
}

func (m *mockWriteCloser) Write(p []byte) (int, error) {
	m.writeCount++
	return m.data.Write(p)
}

func (m *mockWriteCloser) Close() error {
	m.closed = true
	return m.closeErr
}

func TestNewGzipReadCloser(t *testing.T) {
	// Create valid gzipped data
	var gzippedBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&gzippedBuf)
	gzWriter.Write([]byte("test content"))
	gzWriter.Close()
	gzippedData := gzippedBuf.Bytes()

	tests := []testlib.TestCase[[]byte, bool]{
		{
			Name:  "valid_gzip",
			Input: gzippedData,
			Want:  true,
		},
		{
			Name:      "invalid_gzip",
			Input:     []byte("not gzipped content"),
			ShouldErr: true,
		},
		{
			Name:      "empty_data",
			Input:     []byte{},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(data []byte) (bool, error) {
		rc := newMockReadCloser(data)
		gz, err := NewGzipReadCloser(rc)
		if err != nil {
			return false, err
		}
		return gz != nil, nil
	})
}

func TestGzipReadCloser_Read(t *testing.T) {
	// Create gzipped data
	var gzippedBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&gzippedBuf)
	gzWriter.Write([]byte("test content"))
	gzWriter.Close()

	rc := newMockReadCloser(gzippedBuf.Bytes())
	gz, err := NewGzipReadCloser(rc)
	if err != nil {
		t.Fatalf("unexpected error creating GzipReadCloser: %v", err)
	}

	// Read content
	buf := make([]byte, 100)
	n, err := gz.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("unexpected error reading: %v", err)
	}

	if string(buf[:n]) != "test content" {
		t.Fatalf("wanted 'test content' but got '%s'", string(buf[:n]))
	}
}

func TestGzipReadCloser_Close(t *testing.T) {
	// Create gzipped data
	var gzippedBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&gzippedBuf)
	gzWriter.Write([]byte("test"))
	gzWriter.Close()

	rc := newMockReadCloser(gzippedBuf.Bytes())
	gz, err := NewGzipReadCloser(rc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = gz.Close()
	if err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	if !rc.closed {
		t.Fatal("underlying reader was not closed")
	}
}

func TestGzipReadCloser_CloseWithError(t *testing.T) {
	// Create gzipped data
	var gzippedBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&gzippedBuf)
	gzWriter.Write([]byte("test"))
	gzWriter.Close()

	rc := newMockReadCloser(gzippedBuf.Bytes())
	rc.closeErr = errors.New("close error")

	gz, err := NewGzipReadCloser(rc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = gz.Close()
	if err == nil {
		t.Fatal("expected close error but got nil")
	}
}

func TestGzipWriteCloser(t *testing.T) {
	wc := &mockWriteCloser{}
	gz := NewGzipWriteCloser(wc)

	// Write data
	testData := []byte("test content for gzip write")
	n, err := gz.Write(testData)
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if n != len(testData) {
		t.Fatalf("wanted write count %d but got %d", len(testData), n)
	}

	// Close
	err = gz.Close()
	if err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	if !wc.closed {
		t.Fatal("underlying writer was not closed")
	}

	// Verify the data can be decompressed
	gzReader, err := gzip.NewReader(&wc.data)
	if err != nil {
		t.Fatalf("error creating gzip reader: %v", err)
	}
	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("error reading decompressed data: %v", err)
	}

	if string(decompressed) != string(testData) {
		t.Fatalf("decompressed data mismatch: wanted '%s' got '%s'", testData, decompressed)
	}
}

func TestGzipWriteCloser_CloseWithError(t *testing.T) {
	wc := &mockWriteCloser{}
	wc.closeErr = errors.New("close error")
	gz := NewGzipWriteCloser(wc)

	// Write some data first to make gzip valid
	gz.Write([]byte("test"))

	// Close should propagate the error
	err := gz.Close()
	if err == nil {
		t.Fatal("expected close error but got nil")
	}
}
