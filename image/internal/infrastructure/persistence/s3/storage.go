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

func (s *Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		var noSuchKey *s3.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil, ErrObjectNotFound
		}

		return nil, fmt.Errorf("s3 get object: %w", err)
	}

	return out.Body, nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("s3 delete object: %w", err)
	}

	return nil
}
