package user

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Profile struct {
	PK          string  `dynamodbav:"pk"`
	SK          string  `dynamodbav:"sk"`
	Username    string  `dynamodbav:"username"`
	Description string  `dynamodbav:"description"`
	Streak      int64   `dynamodbav:"streak"`
	Score       float64 `dynamodbav:"score"`
}

func (c *Controller) GetProfile(ctx context.Context, id string) (*Profile, bool, error) {
	result, err := c.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(c.table),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", id)},
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

func (c *Controller) PutProfile(ctx context.Context, id string, profile *Profile) error {
	profile.PK = fmt.Sprintf("USER#%s", id)
	profile.SK = "PROFILE"
	item, err := attributevalue.MarshalMap(profile)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	_, err = c.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.table),
		Item:      item,
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (c *Controller) DeleteProfile(ctx context.Context, id string) error {
	_, err := c.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", id)},
			"sk": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}
