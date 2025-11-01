// package email provides an application aware wrapper for the required dynamodb communication on the email model.
package email

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Controller struct {
	client *dynamodb.Client
	table  string
}

func New(client *dynamodb.Client, table string) *Controller {
	return &Controller{client, table}
}
