package zen

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Operator struct {
	region string
	account string
	deleteProtection bool
}

func (o *Operator) Deploy(ctx *pulumi.Context) error {
	tableOutputs, err := o.deployTable(ctx)
	if err!=nil {
		return fmt.Errorf("failed to deploy table: %v", err)
	}
	schedulerOutputs, err := o.deployScheduler(ctx)
	if err!=nil {
		return fmt.Errorf("failed to deploy scheduler: %v", err)
	}
	return nil
}
