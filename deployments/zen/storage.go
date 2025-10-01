package zen

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type storageInput struct {
}

type storageOutput struct {
	BucketArn pulumi.StringOutput
	BucketPolicyArn pulumi.StringOutput
}

func (o *Operator) deployStorage(ctx *pulumi.Context, input *storageInput) (*storageOutput, error) {
	bucket, err := s3.NewBucket(ctx, "storage", &s3.BucketArgs{
		BucketPrefix: pulumi.String("zen-storage-"),
		Region: pulumi.String(o.region),
		ForceDestroy: pulumi.BoolPtr(!o.deleteProtection),
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
				"Resource": "%s/leaderboard/*"
			}]
		}`, bucket.Arn),
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
					"s3:PutObject"
				],
				"Resource": "%s/*"
			}]
		}`, bucket.Arn),
	})
	if err != nil {
		return nil, err
	}
	
	return &storageOutput{
		BucketArn: bucket.Arn,
		BucketPolicyArn: bucketPolicy.Arn,
	}, nil
}
