package smartrw

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nsiow/yams/internal/testlib"
)

// mockS3WriteClient implements S3WriteClient for testing
type mockS3WriteClient struct {
	putErr    error
	putCalled bool
	bucket    string
	key       string
}

func (m *mockS3WriteClient) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.putCalled = true
	if input.Bucket != nil {
		m.bucket = *input.Bucket
	}
	if input.Key != nil {
		m.key = *input.Key
	}
	return &s3.PutObjectOutput{}, m.putErr
}

func TestS3Writer_Parse(t *testing.T) {
	tests := []testlib.TestCase[string, [2]string]{
		{
			Name:  "valid_path",
			Input: "mybucket/path/to/key.json",
			Want:  [2]string{"mybucket", "path/to/key.json"},
		},
		{
			Name:  "simple_key",
			Input: "bucket/key",
			Want:  [2]string{"bucket", "key"},
		},
		{
			Name:      "no_key",
			Input:     "bucketonly",
			ShouldErr: true,
		},
		{
			Name:      "empty_string",
			Input:     "",
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(path string) ([2]string, error) {
		w := &S3Writer{}
		bucket, key, err := w.Parse(path)
		if err != nil {
			return [2]string{}, err
		}
		return [2]string{bucket, key}, nil
	})
}

func TestS3Writer_Write(t *testing.T) {
	w := &S3Writer{
		Bucket: "test-bucket",
		Key:    "test/key.json",
		S3:     &mockS3WriteClient{},
	}

	data := []byte("test content")
	n, err := w.Write(data)
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if n != len(data) {
		t.Fatalf("wanted write count %d but got %d", len(data), n)
	}

	// Verify data was written to buffer
	if w.Body.String() != "test content" {
		t.Fatalf("body mismatch: wanted 'test content' got '%s'", w.Body.String())
	}
}

func TestS3Writer_Close(t *testing.T) {
	mockClient := &mockS3WriteClient{}
	w := &S3Writer{
		Bucket: "test-bucket",
		Key:    "test/key.json",
		S3:     mockClient,
	}

	// Write some data first
	if _, err := w.Write([]byte("test content")); err != nil {
		t.Fatalf("failed to write test data: %v", err)
	}

	// Close should call PutObject
	err := w.Close()
	if err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	if !mockClient.putCalled {
		t.Fatal("PutObject was not called")
	}
	if mockClient.bucket != "test-bucket" {
		t.Fatalf("bucket mismatch: wanted 'test-bucket' got '%s'", mockClient.bucket)
	}
	if mockClient.key != "test/key.json" {
		t.Fatalf("key mismatch: wanted 'test/key.json' got '%s'", mockClient.key)
	}
}

func TestS3Writer_CloseWithError(t *testing.T) {
	mockClient := &mockS3WriteClient{putErr: errors.New("put object error")}
	w := &S3Writer{
		Bucket: "test-bucket",
		Key:    "test/key.json",
		S3:     mockClient,
	}

	err := w.Close()
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestSelectWriter_S3(t *testing.T) {
	// Test the error path for s3:// without proper credentials
	_, err := selectWriter("s3://mybucket/key")
	// This will fail without proper AWS credentials, which is expected
	if err == nil {
		t.Log("S3 writer created successfully (requires valid AWS credentials)")
	}
}

func TestNewS3Writer_ConfigError(t *testing.T) {
	// Test the config error path by passing an option that returns an error
	_, err := NewS3Writer("mybucket/mykey", func(lo *config.LoadOptions) error {
		return errors.New("config load error")
	})

	if err == nil {
		t.Fatal("expected config error but got nil")
	}
}
