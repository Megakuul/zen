// package rating provides an application aware wrapper for the required queue mechanism of the rating change communication.
package rating

import (
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Model struct {
	sqsClient *sqs.Client
	queue     string
}

func New(sqs *sqs.Client, queue string) *Model {
	return &Model{sqs, queue}
}
