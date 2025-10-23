package scheduler

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type BuildInput struct {
	CtxPath   string
	CachePath string
}

type BuildOutput struct {
	Handler pulumi.ArchiveOutput
}

func Build(ctx *pulumi.Context, input *BuildInput) (*BuildOutput, error) {
	outputPath, err := filepath.Abs(input.CachePath)
	if err != nil {
		return nil, err
	}
	command := fmt.Sprintf("go build -o '%s' ./cmd/scheduler/scheduler.go", outputPath)
	build, err := local.NewCommand(ctx, "scheduler", &local.CommandArgs{
		Create:       pulumi.String(command),
		Update:       pulumi.String(command),
		Dir:          pulumi.String(input.CtxPath),
		ArchivePaths: pulumi.ToStringArray([]string{outputPath}),
		Environment: pulumi.ToStringMap(map[string]string{
			"CGO_ENABLED": "0",
			"GOOS":        "linux",
			"GOARCH":      "arm64",
		}),
		Logging: local.LoggingStdoutAndStderr,
		// not rebuilding causes the empty archive to trigger a rebuild of the function deployment.
		// therefore, rebuild is always triggered.
		Triggers: pulumi.ToArray([]any{uuid.New().String()}),
	})
	if err != nil {
		return nil, err
	}
	return &BuildOutput{
		Handler: build.Archive,
	}, nil
}

type DeployInput struct {
	Region         string
	Handler        pulumi.ArchiveOutput
	TableName      pulumi.StringOutput
	TablePolicyArn pulumi.StringOutput
	QueueName      pulumi.StringOutput
	QueuePolicyArn pulumi.StringOutput
}

type DeployOutput struct {
	PublicUrl pulumi.StringOutput
}

func Deploy(ctx *pulumi.Context, input *DeployInput) (*DeployOutput, error) {
	schedulerLogGroup, err := cloudwatch.NewLogGroup(ctx, "scheduler", &cloudwatch.LogGroupArgs{
		Name:            pulumi.String("zen-scheduler"),
		Region:          pulumi.String(input.Region),
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
		Region:        pulumi.StringPtr(input.Region),
		Handler:       pulumi.String("scheduler"),
		Runtime:       lambda.RuntimeCustomAL2023,
		Architectures: pulumi.ToStringArray([]string{"arm64"}),
		MemorySize:    pulumi.IntPtr(128),
		LoggingConfig: &lambda.FunctionLoggingConfigArgs{
			LogGroup:  schedulerLogGroup.Name,
			LogFormat: pulumi.String("Text"),
		},
		Role: schedulerRole.Arn,
		Code: input.Handler,
		Environment: &lambda.FunctionEnvironmentArgs{
			Variables: pulumi.ToStringMapOutput(map[string]pulumi.StringOutput{
				"TABLE": input.TableName,
				"QUEUE": input.QueueName,
			}),
		},
	})
	if err != nil {
		return nil, err
	}

	schedulerUrl, err := lambda.NewFunctionUrl(ctx, "scheduler", &lambda.FunctionUrlArgs{
		FunctionName:      scheduler.Arn,
		InvokeMode:        pulumi.String("BUFFERED"),
		Qualifier:         pulumi.String("$LATEST"),
		Region:            pulumi.String(input.Region),
		AuthorizationType: pulumi.String("NONE"),
	})
	if err != nil {
		return nil, err
	}
	return &DeployOutput{
		PublicUrl: schedulerUrl.FunctionUrl,
	}, nil
}
