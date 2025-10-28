package authentication

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/megakuul/zen/pkg/api/v1/manager/authentication"
)

type Authentication struct {
	dynamoClient  *dynamodb.Client
	dynamoTable   string
	sesClient     *ses.Client
	emailIdentity string
}

func New() *Authentication {
	return &Authentication{}
}

func (s *Authentication) Get(ctx context.Context, r *connect.Request[authentication.GetRequest]) (*connect.Response[authentication.GetResponse], error) {
	blocks := strings.SplitN(r.Msg.Verifier, ":", 2)
	if len(blocks) < 2 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid verifier format; expected '<type>:<value>'"))
	}
	typ, value := blocks[0], blocks[1]
	switch typ {
	case "email":
		result, err := s.dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName: aws.String(s.dynamoTable),
			ExpressionAttributeValues: map[string]dynamotypes.AttributeValue{
				"pk": &dynamotypes.AttributeValueMemberS{Value: fmt.Sprintf("VERIFY#%s", value)},
				"now": &dynamotypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Unix())},
			},
			KeyConditionExpression: aws.String("pk = :pk"),
			FilterExpression: aws.String("expiresAt < :now"), // must be expired
		})
		if err != nil {
			return nil, err
		}
		if result.Count < 1 {
			time.Sleep(time.Second) // immitate email sending
			// pseudo success is returned to avoid users to "scan" for registered emails.
			return &connect.Response[authentication.GetResponse]{
				Msg: &authentication.GetResponse{Token: ""},
			}, nil
		}
		verify := result.Items[0]

		codeChars := "23456789ABCDEFGHJKLMNOPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
		code := strings.Builder{}
		for i := range 8 {
			if i == 3 {
				code.WriteByte('-')
			} else {
				codeCharsIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(codeChars)-1)))
				code.WriteByte(codeChars[codeCharsIdx.Int64()])
			}
		}

		_, err = s.dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(s.dynamoTable),
			Item: map[string]dynamotypes.AttributeValue{
				"pk":        verify["pk"],
				"sk":        &dynamotypes.AttributeValueMemberS{Value: fmt.Sprintf("CODE#%s", code)},
				"expiresAt": &dynamotypes.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Add(5*time.Minute).Unix())},
			},
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}

		_, err = s.sesClient.SendEmail(ctx, &ses.SendEmailInput{
			Destination: &types.Destination{
				ToAddresses: []string{value},
			},
			Source: aws.String(s.emailIdentity),
			Message: &types.Message{
				Subject: &types.Content{
					Data: aws.String("Verification Code"), Charset: aws.String("UTF-8"),
				},
				Body: &types.Body{Text: &types.Content{
					Data: aws.String(fmt.Sprintf(
						"Your application verification code is '%s'", code,
					)),
					Charset: aws.String("UTF-8"),
				}},
			},
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		return &connect.Response[authentication.GetResponse]{
			Msg: &authentication.GetResponse{Token: ""},
		}, nil
	case "code":
		result, err := s.dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName: aws.String(s.dynamoTable),
			IndexName: aws.String("email_gsi"),
			ExpressionAttributeValues: map[string]dynamotypes.AttributeValue{
				"email": &dynamotypes.AttributeValueMemberS{Value: value},
				"code": &dynamotypes.AttributeValueMemberS{Value: value},
			},
			KeyConditionExpression: aws.String("email = :email"),
		})
		if err != nil {
			return nil, err
		}
		if result.Count < 1 {
			time.Sleep(time.Second) // immitate email sending
			// pseudo success is returned to avoid users to "scan" for registered emails.
			return &connect.Response[authentication.GetResponse]{
				Msg: &authentication.GetResponse{Token: ""},
			}, nil
		}

	}
	return nil, nil
}

func (s *Authentication) Revoke(ctx context.Context, r *connect.Request[authentication.RevokeRequest]) (*connect.Response[authentication.RevokeResponse], error) {
	return nil, nil
}
