package user

import (
	"context"
	"fmt"
	"strconv"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Profile struct {
	PK          string  `dynamodbav:"pk"`
	SK          string  `dynamodbav:"sk"`
	Username    string  `dynamodbav:"username,omitempty"`
	Description string  `dynamodbav:"description,omitempty"`
	Leaderboard bool    `dynamodbav:"leaderboard,omitempty"`
	CreatedAt   int64   `dynamodbav:"created_at,omitempty"`
	Streak      int64   `dynamodbav:"streak,omitempty"`
	Score       float64 `dynamodbav:"score,omitempty"`
	MaxStreak   int64   `dynamodbav:"max_streak,omitempty"`
}

func (m *Model) GetProfile(ctx context.Context, sub string) (*Profile, bool, error) {
	result, err := m.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(m.table),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
			":sk": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
		KeyConditionExpression: aws.String("pk = :pk AND sk = :sk"),
	})
	if err != nil {
		return nil, false, connect.NewError(connect.CodeInternal, err)
	} else if len(result.Items) < 1 {
		return nil, false, nil
	}

	profile := &Profile{}
	if err := attributevalue.UnmarshalMap(result.Items[0], profile); err != nil {
		return nil, false, connect.NewError(connect.CodeInternal, err)
	}
	return profile, true, nil
}

func (m *Model) PutProfile(ctx context.Context, sub string, profile *Profile) error {
	profile.PK = fmt.Sprintf("USER#%s", sub)
	profile.SK = "PROFILE"
	item, err := attributevalue.MarshalMap(profile)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	_, err = m.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(m.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(pk)"),
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (m *Model) UpdateProfile(ctx context.Context, sub string, profile *Profile) error {
	_, err := m.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(m.table),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
			"sk": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":username":    &types.AttributeValueMemberS{Value: profile.Username},
			":description": &types.AttributeValueMemberS{Value: profile.Description},
			":leaderboard": &types.AttributeValueMemberBOOL{Value: profile.Leaderboard},
		},
		UpdateExpression:    aws.String("SET username = :username, description = :description, leaderboard = :leaderboard"),
		ConditionExpression: aws.String("attribute_exists(pk)"),
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (m *Model) UpdateProfileRating(ctx context.Context, sub string, ratingChange float64) error {
	streakExpr := "streak = streak + 1"
	if ratingChange < 0 {
		streakExpr = "streak = 0" // reset streak if negative
	}
	result, err := m.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(m.table),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
			"sk": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":change": &types.AttributeValueMemberN{Value: strconv.FormatFloat(ratingChange, 'f', 10, 64)},
		},
		UpdateExpression: aws.String(fmt.Sprint("SET",
			"score = score + :change,",
			streakExpr,
		)),
		ConditionExpression: aws.String("attribute_exists(pk)"),
		ReturnValues:        types.ReturnValueUpdatedNew,
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	// FYI: theoretically this could race and swallow one streak update.
	// practically it has no impact as the streak is a statistical value and
	// is expected to be instable if a user abuses the app by spamming updates so fast that they can race.
	_, err = m.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(m.table),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
			"sk": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":streak": result.Attributes["streak"],
		},
		UpdateExpression:    aws.String("SET max_streak = :streak"),
		ConditionExpression: aws.String(":streak > max_streak"),
	})
	return nil
}

func (m *Model) DeleteProfile(ctx context.Context, sub string) error {
	_, err := m.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(m.table),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
			"sk": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}
