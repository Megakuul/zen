package deploy

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type managerInput struct {
	CodeArchive    pulumi.Archive
	BucketName      pulumi.StringOutput
	BucketPolicyArn pulumi.StringOutput
	TableName      pulumi.StringOutput
	TablePolicyArn pulumi.StringOutput
}

type managerOutput struct {
	PublicUrl pulumi.StringOutput
}

func (o *Operator) deployManager(ctx *pulumi.Context, input *managerInput) (*managerOutput, error) {
	managerLogGroup, err := cloudwatch.NewLogGroup(ctx, "manager", &cloudwatch.LogGroupArgs{
		Name:            pulumi.String("zen-manager"),
		Region:          pulumi.String(o.region),
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
		}),
	})
	if err != nil {
		return nil, err
	}

	manager, err := lambda.NewFunction(ctx, "manager", &lambda.FunctionArgs{
		Name:          pulumi.String("zen-manager"),
		Description:   pulumi.StringPtr("backend responsible for managing administrative tasks"),
		Region:        pulumi.StringPtr(o.region),
		Handler:       pulumi.StringPtr("manager"),
		Runtime:       lambda.RuntimeCustomAL2023,
		Architectures: pulumi.ToStringArray([]string{"x8664"}),
		MemorySize:    pulumi.IntPtr(128),
		LoggingConfig: lambda.FunctionLoggingConfigPtr(&lambda.FunctionLoggingConfigArgs{
			LogGroup:  managerLogGroup.Name,
			LogFormat: pulumi.String("Text"),
		}),
		Role: managerRole.Arn,
		Code: input.CodeArchive,
		Environment: lambda.FunctionEnvironmentPtr(&lambda.FunctionEnvironmentArgs{
			Variables: pulumi.ToStringMapOutput(map[string]pulumi.StringOutput{
				"TABLE": input.TableName,
				"BUCKET": input.BucketName,
			}),
		}),
	})
	if err != nil {
		return nil, err
	}

	managerUrl, err := lambda.NewFunctionUrl(ctx, "manager", &lambda.FunctionUrlArgs{
		FunctionName:      manager.Arn,
		InvokeMode:        pulumi.String("BUFFERED"),
		Qualifier:         pulumi.String("$LATEST"),
		Region:            pulumi.String(o.region),
		AuthorizationType: pulumi.String("NONE"),
	})
	if err != nil {
		return nil, err
	}
	return &managerOutput{
		PublicUrl: managerUrl.FunctionUrl,
	}, nil
}
