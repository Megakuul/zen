package deploy

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sqs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type leaderboardInput struct {
	CodeArchive     pulumi.Archive
	BucketName      pulumi.StringOutput
	BucketPolicyArn pulumi.StringOutput
}

type leaderboardOutput struct {
	QueueName      pulumi.StringOutput
	QueueArn       pulumi.StringOutput
	QueuePolicyArn pulumi.StringOutput
}

func (o *Operator) deployLeaderboard(ctx *pulumi.Context, input *leaderboardInput) (*leaderboardOutput, error) {
	queue, err := sqs.NewQueue(ctx, "leaderboard", &sqs.QueueArgs{
		Name:                     pulumi.String("zen-leaderboard"),
		Region:                   pulumi.String(o.region),
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
		Region:          pulumi.String(o.region),
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
		Name:          pulumi.String("zen-leaderboard"),
		Description:   pulumi.StringPtr("background processor responsible for creating the leaderboard"),
		Region:        pulumi.StringPtr(o.region),
		Handler:       pulumi.StringPtr("leaderboard"),
		Runtime:       lambda.RuntimeCustomAL2023,
		Architectures: pulumi.ToStringArray([]string{"x8664"}),
		MemorySize:    pulumi.IntPtr(512),
		Timeout:       queue.VisibilityTimeoutSeconds, // avoid a function to read the task while another is processing it
		LoggingConfig: lambda.FunctionLoggingConfigPtr(&lambda.FunctionLoggingConfigArgs{
			LogGroup:  leaderboardLogGroup.Name,
			LogFormat: pulumi.String("Text"),
		}),
		Role: leaderboardRole.Arn,
		Code: input.CodeArchive,
		Environment: lambda.FunctionEnvironmentPtr(&lambda.FunctionEnvironmentArgs{
			Variables: pulumi.ToStringMapOutput(map[string]pulumi.StringOutput{
				"QUEUE_NAME":  queue.Name,
				"BUCKET_NAME": input.BucketName,
			}),
		}),
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

	return &leaderboardOutput{
		QueueName:      queue.Name,
		QueueArn:       queue.Arn,
		QueuePolicyArn: queuePushPolicy.Arn,
	}, nil
}
