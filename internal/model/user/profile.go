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
	Username    string  `dynamodbav:"username,omitempty"`
	Description string  `dynamodbav:"description,omitempty"`
	Streak      int64   `dynamodbav:"streak,omitempty"`
	Score       float64 `dynamodbav:"score,omitempty"`
}

func (c *Controller) GetProfile(ctx context.Context, sub string) (*Profile, bool, error) {
	result, err := c.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(c.table),
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

func (c *Controller) PutProfile(ctx context.Context, sub string, profile *Profile) error {
	profile.PK = fmt.Sprintf("USER#%s", sub)
	profile.SK = "PROFILE"
	item, err := attributevalue.MarshalMap(profile)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	_, err = c.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(c.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(pk)"),
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (c *Controller) UpdateProfile(ctx context.Context, sub string, profile *Profile) error {
	_, err := c.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", sub)},
			"sk": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":username":    &types.AttributeValueMemberS{Value: profile.Username},
			":description": &types.AttributeValueMemberS{Value: profile.Description},
		},
		UpdateExpression:    aws.String("SET username = :username, description = :description"),
		ConditionExpression: aws.String("attribute_exists(pk)"),
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (c *Controller) DeleteProfile(ctx context.Context, sub string) error {
	_, err := c.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(c.table),
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
