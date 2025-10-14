package deploy

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type storageInput struct {
}

type storageOutput struct {
	BucketName      pulumi.StringOutput
	BucketDomain    pulumi.StringOutput
	BucketArn       pulumi.StringOutput
	BucketPolicyArn pulumi.StringOutput
}

func (o *Operator) deployStorage(ctx *pulumi.Context, input *storageInput) (*storageOutput, error) {
	bucket, err := s3.NewBucket(ctx, "storage", &s3.BucketArgs{
		BucketPrefix: pulumi.String("zen-storage-"),
		Region:       pulumi.String(o.region),
		ForceDestroy: pulumi.BoolPtr(!o.deleteProtection),
	})
	if err != nil {
		return nil, err
	}

	_, err = s3.NewBucketLifecycleConfigurationV2(ctx, "storage", &s3.BucketLifecycleConfigurationV2Args{
		Bucket: bucket.Bucket,
		Region: bucket.Region,
		Rules: s3.BucketLifecycleConfigurationV2RuleArray{
			s3.BucketLifecycleConfigurationV2RuleArgs{
				Id:     pulumi.String("captcha-cleanup"),
				Prefix: pulumi.String("captcha/"),
				Status: pulumi.String("Enabled"),
				Expiration: s3.BucketLifecycleConfigurationV2RuleExpirationPtr(&s3.BucketLifecycleConfigurationV2RuleExpirationArgs{
					Days:                      pulumi.IntPtr(1),
					ExpiredObjectDeleteMarker: pulumi.BoolPtr(true),
				}),
			},
		},
	})
	if err!=nil {
		return nil, err
	}

	_, err = s3.NewBucketPolicy(ctx, "storage", &s3.BucketPolicyArgs{
		Bucket: bucket.Bucket,
		Region: bucket.Region,
		Policy: pulumi.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {
					"Service": "cloudfront.amazonaws.com"
				},
				"Action": "s3:GetObject",
				"Resource": [
					"%s/web/*",
					"%s/leaderboard/*"
				]
			}]
		}`, bucket.Arn, bucket.Arn),
	})
	if err != nil {
		return nil, err
	}

	// for simplicity there is only one rw policy
	// (there is no use for a seperate readonly policy right now)
	bucketPolicy, err := iam.NewPolicy(ctx, "storage", &iam.PolicyArgs{
		Name: pulumi.String("zen-storage-rw"),
		Policy: pulumi.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Action": [
					"s3:GetObject",
					"s3:PutObject",
					"s3:DeleteObject"
				],
				"Resource": "%s/*"
			}]
		}`, bucket.Arn),
	})
	if err != nil {
		return nil, err
	}

	return &storageOutput{
		BucketName:      bucket.Bucket,
		BucketDomain:    bucket.BucketRegionalDomainName,
		BucketArn:       bucket.Arn,
		BucketPolicyArn: bucketPolicy.Arn,
	}, nil
}
