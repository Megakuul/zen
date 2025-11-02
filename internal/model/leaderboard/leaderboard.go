// package leaderboard provides an application aware wrapper for the required s3 and sqs communication on the leaderboard model.
package leaderboard

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Controller struct {
	sqsClient *sqs.Client
	queue string

	s3Client *s3.Client
	bucket string
	prefix string
}

func New(sqs *sqs.Client, s3 *s3.Client, queue, bucket, prefix string) *Controller {
	return &Controller{sqs, queue, s3, bucket, prefix}
}
