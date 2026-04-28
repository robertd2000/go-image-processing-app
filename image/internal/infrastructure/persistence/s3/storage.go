package s3

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var ErrObjectNotFound = errors.New("object not found")

type Storage struct {
	client *s3.Client
	bucket string
}

func New(client *s3.Client, bucket string) *Storage {
	return &Storage{
		client: client,
		bucket: bucket,
	}
}

func (s *Storage) Put(
	ctx context.Context,
	key string,
	r io.Reader,
	size int64,
	contentType string,
) error {

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          r,
		ContentLength: &size,
		ContentType:   aws.String(contentType),
	})

	if err != nil {
		return fmt.Errorf("s3 put object: %w", err)
	}

	return nil
}
