package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Event struct {
	PK             string  `dynamodbav:"pk"`
	SK             string  `dynamodbav:"sk"`
	Type           int64   `dynamodbav:"type"`
	Name           string  `dynamodbav:"name"`
	StartTime      int64   `dynamodbav:"start_time"`
	StopTime       int64   `dynamodbav:"stop_time"`
	TimerStartTime int64   `dynamodbav:"timer_start_time"`
	TimerStopTime  int64   `dynamodbav:"timer_stop_time"`
	RatingChange   float64 `dynamodbav:"rating_change"`
	Immutable      bool    `dynamodbav:"immutable"`
	Description    string  `dynamodbav:"description"`
	MusicUrl       string  `dynamodbav:"music_url"`
}

func (c *Controller) GetEvent(ctx context.Context, sub, id string) (*Event, bool, error) {
	result, err := c.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(c.table),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
			":sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("EVENT#%s", id)},
		},
		KeyConditionExpression: aws.String("pk = :pk AND sk = :sk"),
	})
	if err != nil {
		return nil, false, connect.NewError(connect.CodeInternal, err)
	} else if len(result.Items) < 1 {
		return nil, false, nil
	}

	event := &Event{}
	if err := attributevalue.UnmarshalMap(result.Items[0], event); err != nil {
		return nil, false, connect.NewError(connect.CodeInternal, err)
	}
	return event, true, nil
}

func (c *Controller) ListEvents(ctx context.Context, sub string, since, until time.Time) ([]*Event, error) {
	result, err := c.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(c.table),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
			":sk_prefix": &types.AttributeValueMemberS{Value: "EVENT#"},
			":since":     &types.AttributeValueMemberS{Value: fmt.Sprintf("EVENT#%d", since.Unix())},
			":until":     &types.AttributeValueMemberS{Value: fmt.Sprintf("EVENT#%d", until.Unix())},
		},
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk) AND sk >= :since AND sk <= :until"),
		ScanIndexForward:       aws.Bool(false),
		Limit:                  aws.Int32(100),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	events := []*Event{}
	for _, item := range result.Items {
		event := &Event{}
		if err := attributevalue.UnmarshalMap(item, event); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		events = append(events, event)
	}
	return events, nil
}

func (c *Controller) PutEvent(ctx context.Context, sub string, event *Event) error {
	event.PK = fmt.Sprintf("USER#%s", sub)
	event.SK = fmt.Sprintf("EVENT#%d", event.StartTime)
	item, err := attributevalue.MarshalMap(event)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	_, err = c.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.table),
		Item:      item,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":now":   &types.AttributeValueMemberN{Value: strconv.Itoa(int(time.Now().Unix()))},
			":false": &types.AttributeValueMemberBOOL{Value: false},
		},
		ConditionExpression: aws.String("stop_time < :now AND immutable = :false"),
	})
	if err != nil {
		var cErr *types.ConditionalCheckFailedException
		if errors.As(err, &cErr) {
			return connect.NewError(connect.CodeOutOfRange, fmt.Errorf("cannot change past or immutable events"))
		}
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (c *Controller) UpdateEventTimer(ctx context.Context, sub, id string, start, stop time.Time, rating float64, immutable bool) error {
	_, err := c.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("EVENT#%s", id)},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":timer_start_time": &types.AttributeValueMemberN{Value: strconv.Itoa(int(start.Unix()))},
			":timer_stop_time":  &types.AttributeValueMemberN{Value: strconv.Itoa(int(stop.Unix()))},
			":rating_change":    &types.AttributeValueMemberN{Value: strconv.FormatFloat(rating, 'f', 2, 64)},
			":immutable":        &types.AttributeValueMemberBOOL{Value: immutable},
			":false":            &types.AttributeValueMemberBOOL{Value: false},
		},
		UpdateExpression:    aws.String("SET timer_start_time = :timer_start_time, timer_stop_time = :timer_stop_time, rating_change = :rating_change"),
		ConditionExpression: aws.String("attribute_exists(sk) AND immutable = :false"),
	})
	if err != nil {
		var cErr *types.ConditionalCheckFailedException
		if errors.As(err, &cErr) {
			return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("event is immutable  or does not exist"))
		}
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (c *Controller) DeleteEvent(ctx context.Context, sub, id string) error {
	_, err := c.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("EVENT#%s", id)},
		},
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}
