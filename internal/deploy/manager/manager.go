package manager

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
	contextPath, err := filepath.Abs(input.CtxPath)
	if err != nil {
		return nil, err
	}
	outputPath, err := filepath.Abs(input.CachePath)
	if err != nil {
		return nil, err
	}
	command := fmt.Sprintf("go build -o '%s' ./cmd/manager/manager.go", outputPath)
	build, err := local.NewCommand(ctx, "manager", &local.CommandArgs{
		Create:       pulumi.String(command),
		Update:       pulumi.String(command),
		Dir:          pulumi.String(contextPath),
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
	TableName       pulumi.StringOutput
	TablePolicyArn  pulumi.StringOutput
	EmailName       pulumi.StringOutput
	EmailPolicyArn  pulumi.StringOutput
}

type DeployOutput struct {
	PublicUrl pulumi.StringOutput
}

func Deploy(ctx *pulumi.Context, input *DeployInput) (*DeployOutput, error) {
	managerLogGroup, err := cloudwatch.NewLogGroup(ctx, "manager", &cloudwatch.LogGroupArgs{
		Name:            pulumi.String("zen-manager"),
		Region:          pulumi.String(input.Region),
		LogGroupClass:   pulumi.String("INFREQUENT_ACCESS"),
		RetentionInDays: pulumi.IntPtr(7),
	})
	if err != nil {
		return nil, err
	}

	managerRole, err := iam.NewRole(ctx, "manager", &iam.RoleArgs{
		Name: pulumi.String("zen-manager"),
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
			input.BucketPolicyArn,
			input.EmailPolicyArn,
		}),
	})
	if err != nil {
		return nil, err
	}

	manager, err := lambda.NewFunction(ctx, "manager", &lambda.FunctionArgs{
		Name:          pulumi.String("zen-manager"),
		Description:   pulumi.StringPtr("backend responsible for managing administrative tasks"),
		Region:        pulumi.StringPtr(input.Region),
		Handler:       pulumi.String("manager"),
		Runtime:       lambda.RuntimeCustomAL2023,
		Architectures: pulumi.ToStringArray([]string{"arm64"}),
		MemorySize:    pulumi.IntPtr(128),
		LoggingConfig: &lambda.FunctionLoggingConfigArgs{
			LogGroup:  managerLogGroup.Name,
			LogFormat: pulumi.String("Text"),
		},
		Role: managerRole.Arn,
		Code: input.Handler,
		Environment: &lambda.FunctionEnvironmentArgs{
			Variables: pulumi.ToStringMapOutput(map[string]pulumi.StringOutput{
				"TABLE":  input.TableName,
				"BUCKET": input.BucketName,
				"EMAIL":  input.EmailName,
			}),
		},
	})
	if err != nil {
		return nil, err
	}

	managerUrl, err := lambda.NewFunctionUrl(ctx, "manager", &lambda.FunctionUrlArgs{
		FunctionName:      manager.Arn,
		InvokeMode:        pulumi.String("BUFFERED"),
		Qualifier:         pulumi.String("$LATEST"),
		Region:            pulumi.String(input.Region),
		AuthorizationType: pulumi.String("NONE"),
	})
	if err != nil {
		return nil, err
	}
	return &DeployOutput{
		PublicUrl: managerUrl.FunctionUrl,
	}, nil
}
