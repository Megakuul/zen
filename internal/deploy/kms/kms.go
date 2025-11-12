package kms

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DeployInput struct {
	Region string
}

type DeployOutput struct {
	KmsName      pulumi.StringOutput
	KmsPolicyArn pulumi.StringOutput
}

func Deploy(ctx *pulumi.Context, input *DeployInput) (*DeployOutput, error) {
	kmsKey, err := kms.NewKey(ctx, "kms", &kms.KeyArgs{
		Region:                pulumi.String(input.Region),
		Description:           pulumi.String("key used as jwt backend for access tokens"),
		KeyUsage:              pulumi.String("SIGN_VERIFY"),
		CustomerMasterKeySpec: pulumi.String("ECC_NIST_P256"),
	})
	if err != nil {
		return nil, err
	}

	kmsPolicy, err := iam.NewPolicy(ctx, "kms", &iam.PolicyArgs{
		Name: pulumi.String("zen-kms-rw"),
		Policy: pulumi.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Action": [
					"kms:Sign",
					"kms:Verify",
					"kms:GetPublicKey"
				],
				"Resource": "%s"
			}]
		}`, kmsKey.Arn),
	})
	if err != nil {
		return nil, err
	}

	return &DeployOutput{
		KmsName:      kmsKey.KeyId,
		KmsPolicyArn: kmsPolicy.Arn,
	}, nil
}
