package email

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Code struct {
	PK        string `dynamodbav:"pk"`
	SK        string `dynamodbav:"sk"`
	Code      string `dynamodbav:"code,omitempty"`
	ExpiresAt int64  `dynamodbav:"expires_at,omitempty"`
}

func (m *Model) GetCode(ctx context.Context, email string) (*Code, bool, error) {
	result, err := m.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(m.table),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":  &types.AttributeValueMemberS{Value: fmt.Sprintf("EMAIL#%s", email)},
			":sk":  &types.AttributeValueMemberS{Value: "CODE"},
			":now": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Unix())},
		},
		KeyConditionExpression: aws.String("pk = :pk AND sk = :sk"),
		FilterExpression:       aws.String("expires_at > :now"),
	})
	if err != nil {
		return nil, false, connect.NewError(connect.CodeInternal, err)
	} else if len(result.Items) < 1 {
		return nil, false, nil
	}

	code := &Code{}
	if err := attributevalue.UnmarshalMap(result.Items[0], code); err != nil {
		return nil, false, connect.NewError(connect.CodeInternal, err)
	}
	return code, true, nil
}

func (m *Model) PutCode(ctx context.Context, email string, code *Code) error {
	code.PK = fmt.Sprintf("EMAIL#%s", email)
	code.SK = "CODE"
	item, err := attributevalue.MarshalMap(code)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	_, err = m.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(m.table),
		Item:      item,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":now": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Unix())},
		},
		ConditionExpression: aws.String("expires_at < :now"),
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}
