package rating

import (
	"context"
	"encoding/json"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Update struct {
	ReceiptHandle string    `json:"-"`
	Deadline      time.Time `json:"-"`
	UserId        string    `json:"user_id"`
	Username      string    `json:"username"`
	Algorithm     string    `json:"algorithm"`
	RatingChange  float64   `json:"rating_change"`
}

func (m *Model) ReadUpdates(ctx context.Context) ([]*Update, error) {
	timeout := 30 * time.Second
	deadline := time.Now().Add(timeout)
	result, err := m.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(m.queue),
		MaxNumberOfMessages: 10,
		VisibilityTimeout:   int32(timeout.Seconds()),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	updates := []*Update{}
	for _, message := range result.Messages {
		update := &Update{
			ReceiptHandle: *message.ReceiptHandle,
			Deadline:      deadline,
		}
		err = json.Unmarshal([]byte(*message.Body), update)
		if err != nil {
			return nil, err
		}
		updates = append(updates, update)
	}
	return updates, nil
}

func (m *Model) DeleteUpdate(ctx context.Context, update *Update) error {
	_, err := m.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(m.queue),
		ReceiptHandle: aws.String(update.ReceiptHandle),
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (m *Model) SendUpdate(ctx context.Context, update *Update) error {
	rawUpdate, err := json.Marshal(update)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	_, err = m.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(m.queue),
		MessageBody: aws.String(string(rawUpdate)),
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}
