// package captcha provides a store implementation for dchest/captcha.Store that uses s3.
package captcha

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Store struct {
	client  *s3.Client
	logger  *slog.Logger
	bucket  string
	prefix  string
	timeout time.Duration
}

func New(client *s3.Client, logger *slog.Logger, timeout time.Duration, bucket, prefix string) *Store {
	return &Store{
		client:  client,
		logger:  logger,
		bucket:  bucket,
		prefix:  prefix,
		timeout: timeout,
	}
}

func (s *Store) Get(id string, del bool) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fmt.Sprint(s.prefix, id)),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if !errors.As(err, &nsk) {
			s.logger.Error(fmt.Sprintf("failed to load captcha: %v", err))
		}
		return nil
	}
	digits, err := io.ReadAll(output.Body)
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to read captcha: %v", err))
		return nil
	}
	if del {
		_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(fmt.Sprint(s.prefix, id)),
		})
		if err != nil {
			s.logger.Error(fmt.Sprintf("failed to delete captcha: %v", err))
			return nil
		}
	}
	return digits
}

func (s *Store) Set(id string, digits []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fmt.Sprint(s.prefix, id)),
		Body:   bytes.NewReader(digits),
	})
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to insert captcha: %v", err))
	}
}
