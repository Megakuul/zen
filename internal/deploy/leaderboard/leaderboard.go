package leaderboard

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sqs"
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
	command := fmt.Sprintf("go build -o '%s' ./cmd/leaderboard/leaderboard.go", outputPath)
	build, err := local.NewCommand(ctx, "leaderboard", &local.CommandArgs{
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
	Region          string
	Handler         pulumi.ArchiveOutput
	BucketName      pulumi.StringOutput
	BucketPolicyArn pulumi.StringOutput
}

type DeployOutput struct {
	QueueName      pulumi.StringOutput
	QueueArn       pulumi.StringOutput
	QueuePolicyArn pulumi.StringOutput
}

func Deploy(ctx *pulumi.Context, input *DeployInput) (*DeployOutput, error) {
	queue, err := sqs.NewQueue(ctx, "leaderboard", &sqs.QueueArgs{
		Name:                     pulumi.String("zen-leaderboard"),
		Region:                   pulumi.String(input.Region),
		MessageRetentionSeconds:  pulumi.IntPtr(1209600), // 14 days -> max
		VisibilityTimeoutSeconds: pulumi.IntPtr(120),     // this also defines the lambda timeout
	})
	if err != nil {
		return nil, err
	}

	queuePullPolicy, err := iam.NewPolicy(ctx, "queue-pull", &iam.PolicyArgs{
		Name: pulumi.String("zen-queue-pull"),
		Policy: pulumi.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Action": [
					"sqs:ReceiveMessage",
					"sqs:DeleteMessage",
					"sqs:ChangeMessageVisibility",
					"sqs:GetQueueAttributes",
					"sqs:GetQueueUrl"
				],
				"Resource": "%s"
			}]
		}`, queue.Arn),
	})
	if err != nil {
		return nil, err
	}

	queuePushPolicy, err := iam.NewPolicy(ctx, "queue-push", &iam.PolicyArgs{
		Name: pulumi.String("zen-queue-push"),
		Policy: pulumi.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Action": [
					"sqs:SendMessage",
					"sqs:SendMessageBatch",
					"sqs:GetQueueAttributes",
					"sqs:GetQueueUrl"
				],
				"Resource": "%s"
			}]
		}`, queue.Arn),
	})
	if err != nil {
		return nil, err
	}

	leaderboardLogGroup, err := cloudwatch.NewLogGroup(ctx, "leaderboard", &cloudwatch.LogGroupArgs{
		Name:            pulumi.String("zen-leaderboard"),
		Region:          pulumi.String(input.Region),
		LogGroupClass:   pulumi.String("INFREQUENT_ACCESS"),
		RetentionInDays: pulumi.IntPtr(7),
	})
	if err != nil {
		return nil, err
	}

	leaderboardRole, err := iam.NewRole(ctx, "leaderboard", &iam.RoleArgs{
		Name: pulumi.String("zen-leaderboard"),
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
			input.BucketPolicyArn,
			queuePullPolicy.Arn,
		}),
	})
	if err != nil {
		return nil, err
	}

	leaderboard, err := lambda.NewFunction(ctx, "leaderboard", &lambda.FunctionArgs{
		Name:        pulumi.String("zen-leaderboard"),
		Description: pulumi.StringPtr("background processor responsible for creating the leaderboard"),
		Region:      pulumi.StringPtr(input.Region),
		Handler: input.Handler.ApplyT(func(archive pulumi.Archive) string {
			return filepath.Base(archive.Path())
		}).(pulumi.StringOutput).ToStringPtrOutput(),
		Runtime:       lambda.RuntimeCustomAL2023,
		Architectures: pulumi.ToStringArray([]string{"arm64"}),
		MemorySize:    pulumi.IntPtr(512),
		Timeout:       queue.VisibilityTimeoutSeconds, // avoid a function to read the task while another is processing it
		LoggingConfig: lambda.FunctionLoggingConfigArgs{
			LogGroup:  leaderboardLogGroup.Name,
			LogFormat: pulumi.String("Text"),
		},
		Role: leaderboardRole.Arn,
		Code: input.Handler,
		Environment: &lambda.FunctionEnvironmentArgs{
			Variables: pulumi.ToStringMapOutput(map[string]pulumi.StringOutput{
				"QUEUE":  queue.Name,
				"BUCKET": input.BucketName,
			}),
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = lambda.NewEventSourceMapping(ctx, "leaderboard", &lambda.EventSourceMappingArgs{
		FunctionName:                   leaderboard.Arn,
		EventSourceArn:                 queue.Arn,
		MaximumRetryAttempts:           pulumi.IntPtr(1),
		BatchSize:                      pulumi.IntPtr(10),
		MaximumBatchingWindowInSeconds: pulumi.IntPtr(60),
	})
	if err != nil {
		return nil, err
	}

	return &DeployOutput{
		QueueName:      queue.Name,
		QueueArn:       queue.Arn,
		QueuePolicyArn: queuePushPolicy.Arn,
	}, nil
}
