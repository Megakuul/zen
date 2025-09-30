package zen

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type schedulerOutputs struct {
	PublicUrl pulumi.StringOutput
}

func (o *Operator) deployScheduler(ctx *pulumi.Context) (*schedulerOutputs, error) {
	schedulerLogGroup, err := cloudwatch.NewLogGroup(ctx, "scheduler", &cloudwatch.LogGroupArgs{
		Name:            pulumi.String("zen-scheduler"),
		Region:          pulumi.String(o.region),
		LogGroupClass:   pulumi.String("INFREQUENT_ACCESS"),
		RetentionInDays: pulumi.IntPtr(7),
	})
	if err != nil {
		return nil, err
	}

	tableRwPolicy, err := iam.NewPolicy(ctx, "table", &iam.PolicyArgs{
		Name: pulumi.String("zen-table-rw"),
		Policy: pulumi.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Action": [
					"dynamodb:GetItem",
					"dynamodb:Query",
					"dynamodb:Scan",
					"dynamodb:PutItem",
					"dynamodb:UpdateItem"
				],
				"Resource": "arn:aws:dynamodb:%s:%s:table/zen-table"
			}]
		}`, o.region, o.account),
	})
	if err != nil {
		return nil, err
	}

	schedulerRole, err := iam.NewRole(ctx, "scheduler", &iam.RoleArgs{
		Name: pulumi.String("zen-scheduler"),
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {
					"Service": "lambda.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
			}]
		}`),
		ManagedPolicyArns: pulumi.ToStringArrayOutput([]pulumi.StringOutput{
			tableRwPolicy.Arn,
		}),
	})
	if err != nil {
		return nil, err
	}

	scheduler, err := lambda.NewFunction(ctx, "scheduler", &lambda.FunctionArgs{
		Name:          pulumi.String("zen-scheduler"),
		Description:   pulumi.StringPtr("backend responsible for managing the calendar associated timings"),
		Region:        pulumi.StringPtr(o.region),
		Handler:       pulumi.StringPtr("scheduler"),
		Runtime:       lambda.RuntimeCustomAL2023,
		Architectures: pulumi.ToStringArray([]string{"x8664"}),
		MemorySize:    pulumi.IntPtr(128),
		LoggingConfig: lambda.FunctionLoggingConfigPtr(&lambda.FunctionLoggingConfigArgs{
			LogGroup:  schedulerLogGroup.Name,
			LogFormat: pulumi.String("Text"),
		}),
		Role: schedulerRole.Arn,
	})
	if err != nil {
		return nil, err
	}

	schedulerUrl, err := lambda.NewFunctionUrl(ctx, "scheduler", &lambda.FunctionUrlArgs{
		FunctionName: scheduler.Arn,
		InvokeMode:   pulumi.String("BUFFERED"),
		Cors: lambda.FunctionUrlCorsPtr(&lambda.FunctionUrlCorsArgs{
			AllowOrigins: pulumi.ToStringArray([]string{"*"}),
			AllowMethods: pulumi.ToStringArray([]string{"GET", "POST", "OPTIONS"}),
			AllowHeaders: pulumi.ToStringArray([]string{
				"Content-Type", "Accept", "Authorization",
				"Connect-Protocol-Version", "Connect-Timeout-Ms",
				"Connect-Accept-Encoding", "Connect-Content-Encoding",
			}),
			ExposeHeaders: pulumi.ToStringArray([]string{
				"Content-Type", "Connect-Protocol-Version",
				"Connect-Accept-Encoding", "Connect-Content-Encoding",
			}),
			AllowCredentials: pulumi.BoolPtr(false),
		}),
		Qualifier:         pulumi.String("$LATEST"),
		Region:            pulumi.String(o.region),
		AuthorizationType: pulumi.String("NONE"),
	})
	return &schedulerOutputs{
		PublicUrl: schedulerUrl.FunctionUrl, 
	}, err
}
