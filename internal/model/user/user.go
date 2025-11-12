// package user provides an application aware wrapper for the required dynamodb communication on the user model.
package user

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Model struct {
	client *dynamodb.Client
	table  string
}

func New(client *dynamodb.Client, table string) *Model {
	return &Model{client, table}
}
