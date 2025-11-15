// package leaderboard provides an application aware wrapper for the required s3 communication on the leaderboard model.
package leaderboard

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Model struct {
	s3Client *s3.Client
	bucket   string
	prefix   string
}

func New(s3 *s3.Client, bucket, prefix string) *Model {
	return &Model{s3, bucket, prefix}
}
