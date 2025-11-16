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
	Time         time.Time `json:"time"`
	UserId       string    `json:"user_id"`
	Username     string    `json:"username"`
	Streak       int64     `json:"streak"`
	Algorithm    string    `json:"algorithm"`
	RatingChange float64   `json:"rating_change"`
}

func (m *Model) ParseUpdate(body string) (*Update, error) {
	update := &Update{}
	err := json.Unmarshal([]byte(body), update)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return update, nil
}

// Caution: if you do not plan to find shelter under a bridge,
// consider NEVER calling this on the leaderboard function triggered by the update...
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
