package deploy

import (
	"path/filepath"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type schedulerInput struct {
	Handler        pulumi.Archive
	TableName      pulumi.StringOutput
	TablePolicyArn pulumi.StringOutput
	QueueName      pulumi.StringOutput
	QueuePolicyArn pulumi.StringOutput
}

type schedulerOutput struct {
	PublicUrl pulumi.StringOutput
}

func (o *Operator) deployScheduler(ctx *pulumi.Context, input *schedulerInput) (*schedulerOutput, error) {
	schedulerLogGroup, err := cloudwatch.NewLogGroup(ctx, "scheduler", &cloudwatch.LogGroupArgs{
		Name:            pulumi.String("zen-scheduler"),
		Region:          pulumi.String(o.region),
		LogGroupClass:   pulumi.String("INFREQUENT_ACCESS"),
		RetentionInDays: pulumi.IntPtr(7),
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
			input.TablePolicyArn,
			input.QueuePolicyArn,
		}),
	})
	if err != nil {
		return nil, err
	}

	scheduler, err := lambda.NewFunction(ctx, "scheduler", &lambda.FunctionArgs{
		Name:          pulumi.String("zen-scheduler"),
		Description:   pulumi.StringPtr("backend responsible for managing the calendar associated timings"),
		Region:        pulumi.StringPtr(o.region),
		Handler:       pulumi.StringPtr(filepath.Base(input.Handler.Path())),
		Runtime:       lambda.RuntimeCustomAL2023,
		Architectures: pulumi.ToStringArray([]string{"arm64"}),
		MemorySize:    pulumi.IntPtr(128),
		LoggingConfig: lambda.FunctionLoggingConfigPtr(&lambda.FunctionLoggingConfigArgs{
			LogGroup:  schedulerLogGroup.Name,
			LogFormat: pulumi.String("Text"),
		}),
		Role: schedulerRole.Arn,
		Code: input.Handler,
		Environment: lambda.FunctionEnvironmentPtr(&lambda.FunctionEnvironmentArgs{
			Variables: pulumi.ToStringMapOutput(map[string]pulumi.StringOutput{
				"TABLE": input.TableName,
				"QUEUE": input.QueueName,
			}),
		}),
	})
	if err != nil {
		return nil, err
	}

	schedulerUrl, err := lambda.NewFunctionUrl(ctx, "scheduler", &lambda.FunctionUrlArgs{
		FunctionName:      scheduler.Arn,
		InvokeMode:        pulumi.String("BUFFERED"),
		Qualifier:         pulumi.String("$LATEST"),
		Region:            pulumi.String(o.region),
		AuthorizationType: pulumi.String("NONE"),
	})
	if err != nil {
		return nil, err
	}
	return &schedulerOutput{
		PublicUrl: schedulerUrl.FunctionUrl,
	}, nil
}
