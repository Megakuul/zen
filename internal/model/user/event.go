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
	PK              string  `dynamodbav:"pk"`
	SK              string  `dynamodbav:"sk"`
	Type            int64   `dynamodbav:"type"`
	Name            string  `dynamodbav:"name"`
	StartTime       int64   `dynamodbav:"start_time"`
	StopTime        int64   `dynamodbav:"stop_time"`
	TimerStartTime  int64   `dynamodbav:"timer_start_time"`
	TimerStopTime   int64   `dynamodbav:"timer_stop_time"`
	RatingChange    float64 `dynamodbav:"rating_change"`
	RatingAlgorithm string  `dynamodbav:"rating_algorithm"`
	Immutable       bool    `dynamodbav:"immutable"`
	Description     string  `dynamodbav:"description"`
	MusicUrl        string  `dynamodbav:"music_url"`
}

func (m *Model) GetEvent(ctx context.Context, sub, id string) (*Event, bool, error) {
	result, err := m.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(m.table),
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

func (m *Model) ListEvents(ctx context.Context, sub string, since, until time.Time) ([]*Event, error) {
	result, err := m.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(m.table),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":    &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
			":since": &types.AttributeValueMemberS{Value: fmt.Sprintf("EVENT#%d", since.Unix())},
			":until": &types.AttributeValueMemberS{Value: fmt.Sprintf("EVENT#%d", until.Unix())},
		},
		KeyConditionExpression: aws.String("pk = :pk AND sk BETWEEN :since AND :until"),
		ScanIndexForward:       aws.Bool(true),
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

// PutEvents inserts all provided events and deletes all specified old events in an all or nothing operation.
// It ensures that events colliding with oldEvents are not deleted (insert: "events" delete: "oldEvents - events").
func (m *Model) PutEvents(ctx context.Context, sub string, events []Event, oldEvents map[string]bool) error {
	writes := []types.TransactWriteItem{}
	for _, event := range events {
		newId := fmt.Sprintf("%d", event.StartTime)
		if oldEvents[newId] {
			delete(oldEvents, newId)
		}
		event.PK = fmt.Sprintf("USER#%s", sub)
		event.SK = fmt.Sprintf("EVENT#%s", newId)
		item, err := attributevalue.MarshalMap(event)
		if err != nil {
			return connect.NewError(connect.CodeInvalidArgument, err)
		}
		writes = append(writes, types.TransactWriteItem{
			Put: &types.Put{
				TableName: aws.String(m.table),
				Item:      item,
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":false": &types.AttributeValueMemberBOOL{Value: false},
				},
				ConditionExpression: aws.String("attribute_not_exists(pk) OR immutable = :false"),
			},
		})
	}

	for id, ok := range oldEvents {
		if !ok || id == "" {
			continue
		}
		writes = append(writes, types.TransactWriteItem{
			Delete: &types.Delete{
				TableName: aws.String(m.table),
				Key: map[string]types.AttributeValue{
					"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
					"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("EVENT#%s", id)},
				},
			},
		})
	}
	_, err := m.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: writes,
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

func (m *Model) UpdateEventTimer(ctx context.Context, sub, id string, start, stop time.Time, rating float64, ratingAlgorithm string, immutable bool) error {
	_, err := m.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(m.table),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("EVENT#%s", id)},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":timer_start_time": &types.AttributeValueMemberN{Value: strconv.Itoa(int(start.Unix()))},
			":timer_stop_time":  &types.AttributeValueMemberN{Value: strconv.Itoa(int(stop.Unix()))},
			":rating_change":    &types.AttributeValueMemberN{Value: strconv.FormatFloat(rating, 'f', 2, 64)},
			":rating_algorithm": &types.AttributeValueMemberS{Value: ratingAlgorithm},
			":immutable":        &types.AttributeValueMemberBOOL{Value: immutable},
			":false":            &types.AttributeValueMemberBOOL{Value: false},
		},
		UpdateExpression: aws.String(fmt.Sprint("SET ",
			"timer_start_time = :timer_start_time,",
			"timer_stop_time = :timer_stop_time,",
			"rating_change = :rating_change,",
			"rating_algorithm = :rating_algorithm,",
			"immutable = :immutable",
		)),
		ConditionExpression: aws.String("attribute_exists(sk) AND immutable = :false"),
	})
	if err != nil {
		var cErr *types.ConditionalCheckFailedException
		if errors.As(err, &cErr) {
			return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("event is immutable or does not exist"))
		}
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (m *Model) DeleteEvent(ctx context.Context, sub, id string) error {
	_, err := m.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(m.table),
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
