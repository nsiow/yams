package smartrw

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3WriteClient allows for swappable S3 implementations where needed
//
// Primarily useful for testing
type S3WriteClient interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type S3Writer struct {
	Bucket string
	Key    string
	Body   bytes.Buffer
	S3     S3WriteClient
}

func NewS3Writer(s3path string, opts ...func(*config.LoadOptions) error) (*S3Writer, error) {
	w := S3Writer{}
	var err error

	w.Bucket, w.Key, err = w.Parse(s3path)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return nil, err
	}
	w.S3 = s3.NewFromConfig(cfg)

	return &w, nil
}

func (s *S3Writer) Parse(src string) (string, string, error) {
	components := strings.SplitN(src, "/", 2)
	if len(components) != 2 {
		return "", "", fmt.Errorf("invalid s3 path: %s", src)
	}

	return components[0], components[1], nil
}

func (s *S3Writer) Write(p []byte) (n int, err error) {
	return s.Body.Write(p)
}

func (s *S3Writer) Close() error {
	_, err := s.S3.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &s.Bucket,
		Key:    &s.Key,
		Body:   &s.Body,
	})

	return err
}
