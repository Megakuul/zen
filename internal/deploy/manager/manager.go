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
	cachePath := ".cache/lambda"
	outputPath := "manager"
	command := fmt.Sprintf("go build -o %s ../../cmd/manager/manager.go", outputPath)
	build, err := local.NewCommand(ctx, "manager", &local.CommandArgs{
		Create: pulumi.String(command),
		Update: pulumi.String(command),
		// must be inside cache otherwise the output archive contains cache paths
		Dir:          pulumi.String(filepath.Join(contextPath, cachePath)),
		ArchivePaths: pulumi.ToStringArray([]string{outputPath}),
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
	Region          string
	Domain          string
	Issuer          string
	Handler         pulumi.ArchiveOutput
	BucketName      pulumi.StringOutput
	BucketPolicyArn pulumi.StringOutput
	TableName       pulumi.StringOutput
	TablePolicyArn  pulumi.StringOutput
	KmsName         pulumi.StringOutput
	KmsPolicyArn    pulumi.StringOutput
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
			input.KmsPolicyArn,
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
				"TABLE":                 input.TableName,
				"TOKEN_ISSUER":          pulumi.Sprintf(input.Issuer),
				"TOKEN_KMS_KEY_ID":      input.KmsName,
				"AUTH_MAIL_SENDER":      input.EmailName,
				"CAPTCHA_BUCKET":        input.BucketName,
				"CAPTCHA_BUCKET_PREFIX": pulumi.Sprintf("captcha/"),
			}),
		},
	})
	if err != nil {
		return nil, err
	}

	managerUrl, err := lambda.NewFunctionUrl(ctx, "manager", &lambda.FunctionUrlArgs{
		FunctionName:      manager.Arn,
		InvokeMode:        pulumi.String("BUFFERED"),
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
