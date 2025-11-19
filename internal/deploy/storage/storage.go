package storage

import (
	"fmt"
	"mime"
	"path"
	"path/filepath"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DeployInput struct {
	Region           string
	DeleteProtection bool
	WebArtifacts     pulumi.AssetOrArchiveMapOutput
}

type DeployOutput struct {
	BucketName      pulumi.StringOutput
	BucketDomain    pulumi.StringOutput
	BucketArn       pulumi.StringOutput
	BucketPolicyArn pulumi.StringOutput
}

func Deploy(ctx *pulumi.Context, input *DeployInput) (*DeployOutput, error) {
	bucket, err := s3.NewBucket(ctx, "storage", &s3.BucketArgs{
		BucketPrefix: pulumi.String("zen-storage-"),
		Region:       pulumi.String(input.Region),
		ForceDestroy: pulumi.BoolPtr(!input.DeleteProtection),
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
				Expiration: &s3.BucketLifecycleConfigurationV2RuleExpirationArgs{
					Days:                      pulumi.IntPtr(1),
					ExpiredObjectDeleteMarker: pulumi.BoolPtr(true),
				},
			},
			s3.BucketLifecycleConfigurationV2RuleArgs{
				Id:     pulumi.String("leaderboard-cleanup"),
				Prefix: pulumi.String("leaderboard/"),
				Status: pulumi.String("Enabled"),
				Expiration: &s3.BucketLifecycleConfigurationV2RuleExpirationArgs{
					Days:                      pulumi.IntPtr(365),
					ExpiredObjectDeleteMarker: pulumi.BoolPtr(true),
				},
			},
		},
	})
	if err != nil {
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

	input.WebArtifacts.ApplyT(func(artifacts map[string]pulumi.AssetOrArchive) (string, error) {
		for key, artifact := range artifacts {
			if asset, ok := artifact.(pulumi.Asset); ok {
				_, err := s3.NewBucketObjectv2(ctx, fmt.Sprintf("storage-web-%s", key), &s3.BucketObjectv2Args{
					Bucket:      bucket.Bucket,
					Key:         pulumi.Sprintf(path.Join("web", key)),
					ContentType: pulumi.String(mime.TypeByExtension(filepath.Ext(asset.Path()))),
					Source:      asset,
				})
				if err != nil {
					return "", err
				}
			}
		}

		return "", nil
	})

	return &DeployOutput{
		BucketName:      bucket.Bucket,
		BucketDomain:    bucket.BucketRegionalDomainName,
		BucketArn:       bucket.Arn,
		BucketPolicyArn: bucketPolicy.Arn,
	}, nil
}
