package scheduler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type BuildInput struct {
	CtxPath string
}

type BuildOutput struct {
	Handler pulumi.ArchiveOutput
}

func Build(ctx *pulumi.Context, input *BuildInput) (*BuildOutput, error) {
	contextPath, err := filepath.Abs(input.CtxPath)
	if err != nil {
		return nil, err
	}
	commandPath := filepath.Join(contextPath, ".cache/scheduler")
	if err = os.MkdirAll(commandPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache path: %v", err)
	}
	command := "go build -o bootstrap ../../cmd/scheduler/scheduler.go"
	build, err := local.NewCommand(ctx, "scheduler", &local.CommandArgs{
		Create: pulumi.String(command),
		Update: pulumi.String(command),
		// must be inside cache otherwise the output archive contains cache paths
		Dir:          pulumi.String(commandPath),
		ArchivePaths: pulumi.ToStringArray([]string{"bootstrap"}),
		Environment: pulumi.ToStringMap(map[string]string{
			"CGO_ENABLED": "0",
			"GOOS":        "linux",
			"GOARCH":      "arm64",
		}),
		Logging: local.LoggingStderr,
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
	Issuer         string
	TableName      pulumi.StringOutput
	TablePolicyArn pulumi.StringOutput
	QueueName      pulumi.StringOutput
	QueuePolicyArn pulumi.StringOutput
	KmsName        pulumi.StringOutput
	KmsPolicyArn   pulumi.StringOutput
}

type DeployOutput struct {
	PublicUrl pulumi.StringOutput
}

func Deploy(ctx *pulumi.Context, input *DeployInput) (*DeployOutput, error) {
	schedulerLogGroup, err := cloudwatch.NewLogGroup(ctx, "scheduler", &cloudwatch.LogGroupArgs{
		Name:            pulumi.String("zen-scheduler"),
		Region:          pulumi.String(input.Region),
		LogGroupClass:   pulumi.String("STANDARD"),
		RetentionInDays: pulumi.IntPtr(7),
	})
	if err != nil {
		return nil, err
	}

	schedulerLogPolicy, err := iam.NewPolicy(ctx, "scheduler-log", &iam.PolicyArgs{
		Name: pulumi.String("zen-scheduler-log-emit"),
		Policy: pulumi.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Action": [
					"logs:CreateLogStream",
					"logs:PutLogEvents"
				],
				"Resource": [
					"%s",
					"%s:log-stream:*"
				]
			}]
		}`, schedulerLogGroup.Arn, schedulerLogGroup.Arn),
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
			schedulerLogPolicy.Arn,
			input.KmsPolicyArn,
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
		Handler:       pulumi.String("bootstrap"),
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
				"TABLE":             input.TableName,
				"TOKEN_ISSUER":      pulumi.Sprintf(input.Issuer),
				"TOKEN_KMS_KEY_ID":  input.KmsName,
				"LEADERBOARD_QUEUE": input.QueueName,
				"RATING_ANCHOR":     pulumi.Sprintf("10m"),
			}),
		},
	})
	if err != nil {
		return nil, err
	}

	schedulerUrl, err := lambda.NewFunctionUrl(ctx, "scheduler", &lambda.FunctionUrlArgs{
		FunctionName:      scheduler.Arn,
		InvokeMode:        pulumi.String("BUFFERED"),
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
