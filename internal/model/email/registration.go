package email

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Registration struct {
	PK   string `dynamodbav:"pk"`
	SK   string `dynamodbav:"sk"`
	User string `dynamodbav:"user"`
}

func (c *Controller) GetRegistration(ctx context.Context, email string) (*Registration, bool, error) {
	result, err := c.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(c.table),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("EMAIL#%s", email)},
			":sk": &types.AttributeValueMemberS{Value: "REGISTRATION"},
		},
		KeyConditionExpression: aws.String("pk = :pk AND sk = :sk"),
	})
	if err != nil {
		return nil, false, connect.NewError(connect.CodeInternal, err)
	} else if len(result.Items) < 1 {
		return nil, false, nil
	}

	registration := &Registration{}
	if err := attributevalue.UnmarshalMap(result.Items[0], registration); err != nil {
		return nil, false, connect.NewError(connect.CodeInternal, err)
	}
	return registration, true, nil
}

func (c *Controller) PutRegistration(ctx context.Context, email string, registration *Registration) error {
	registration.PK = fmt.Sprintf("EMAIL#%s", email)
	registration.SK = "REGISTRATION"
	item, err := attributevalue.MarshalMap(registration)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	_, err = c.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(c.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(pk)"),
	})
	if err!=nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (c *Controller) DeleteRegistration(ctx context.Context, email string) error {
	_, err := c.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("EMAIL#%s", email)},
			"sk": &types.AttributeValueMemberS{Value: "REGISTRATION"},
		},
	})
	if err!=nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}
