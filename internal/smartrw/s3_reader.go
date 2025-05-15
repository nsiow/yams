package smartrw

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// GetObjectClient allows for swappable S3 implementations where needed
//
// Primarily useful for testing
type S3Client interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type S3Reader struct {
	Bucket string
	Key    string
	Body   io.ReadCloser
	S3     S3Client
}

func NewS3Reader(s3path string, opts ...func(*config.LoadOptions) error) (*S3Reader, error) {
	r := S3Reader{}
	var err error

	r.Bucket, r.Key, err = r.Parse(s3path)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return nil, err
	}
	r.S3 = s3.NewFromConfig(cfg)

	return &r, nil
}

func (s *S3Reader) Parse(src string) (string, string, error) {
	components := strings.SplitN(src, "/", 2)
	if len(components) != 2 {
		return "", "", fmt.Errorf("invalid s3 path: %s", src)
	}

	return components[0], components[1], nil
}

func (s *S3Reader) Open() error {
	resp, err := s.S3.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &s.Key,
	})
	if err != nil {
		return err
	}

	s.Body = resp.Body
	return nil
}

func (s *S3Reader) Read(p []byte) (n int, err error) {
	if s.Body == nil {
		err := s.Open()
		if err != nil {
			return 0, err
		}
	}

	return s.Body.Read(p)
}

func (s *S3Reader) Close() error {
	if s.Body != nil {
		return s.Body.Close()
	}

	return nil
}
