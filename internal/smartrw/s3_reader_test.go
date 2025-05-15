package smartrw

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nsiow/yams/internal/testlib"
)

type DummyS3Client struct{}

func (d *DummyS3Client) GetObject(
	ctx context.Context,
	input *s3.GetObjectInput,
	_ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	// Use the "key" argument as filename input
	fp := "../../testdata/smartrw/" + *input.Key
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}

	return &s3.GetObjectOutput{
		Body: f,
	}, nil
}

func TestS3Reader(t *testing.T) {
	tests := []testlib.TestCase[string, string]{
		{
			Name:  "simple_file_1",
			Input: "whateverbucket/test_file_1.json",
			Want:  "892d22932a0e8b5223f860c860238542",
		},
		{
			Name:  "simple_file_2",
			Input: "whateverbucket/test_file_2.json",
			Want:  "abb09e526478388803d74099f0ec6013",
		},
		{
			Name:      "bad_file",
			Input:     "whateverbucket/does_not_exist.json",
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(fp string) (string, error) {
		r, err := NewS3Reader(fp)
		if err != nil {
			return "", err
		}
		defer r.Close()

		// replace S3 implementation with one for unit tests
		r.S3 = &DummyS3Client{}

		data, err := io.ReadAll(r)
		if err != nil {
			return "", err
		}

		hash := md5.Sum(data)
		return fmt.Sprintf("%x", hash), nil
	})
}

func TestBadS3Reader(t *testing.T) {
	_, err := NewS3Reader("bucket/key", config.WithSharedConfigProfile("doesnotexist"))
	if err == nil {
		t.Fatalf("config loading should have failed, but somehow succeeded")
	}
}
