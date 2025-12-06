package table

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DeployInput struct {
	Region           string
	DeleteProtection bool
}

type DeployOutput struct {
	TableName      pulumi.StringOutput
	TablePolicyArn pulumi.StringOutput
}

func Deploy(ctx *pulumi.Context, input *DeployInput) (*DeployOutput, error) {
	table, err := dynamodb.NewTable(ctx, "table", &dynamodb.TableArgs{
		Name:        pulumi.String("zen-table"),
		Region:      pulumi.String(input.Region),
		BillingMode: pulumi.String("PAY_PER_REQUEST"),
		HashKey:     pulumi.String("pk"),
		RangeKey:    pulumi.String("sk"),
		Attributes: dynamodb.TableAttributeArray{
			dynamodb.TableAttributeArgs{Name: pulumi.String("pk"), Type: pulumi.String("S")},
			dynamodb.TableAttributeArgs{Name: pulumi.String("sk"), Type: pulumi.String("S")},
		},
		OnDemandThroughput: &dynamodb.TableOnDemandThroughputArgs{
			MaxWriteRequestUnits: pulumi.IntPtr(10),
			MaxReadRequestUnits:  pulumi.IntPtr(100),
		},
		Ttl: &dynamodb.TableTtlArgs{
			Enabled:       pulumi.BoolPtr(true),
			AttributeName: pulumi.String("expires_at"),
		},
		DeletionProtectionEnabled: pulumi.BoolPtr(input.DeleteProtection),
	})
	if err != nil {
		return nil, err
	}

	// for simplicity there is only one rw policy
	// (there is no use for a seperate readonly policy right now)
	tablePolicy, err := iam.NewPolicy(ctx, "table", &iam.PolicyArgs{
		Name: pulumi.String("zen-table-rw"),
		Policy: pulumi.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Action": [
					"dynamodb:GetItem",
					"dynamodb:Query",
					"dynamodb:PutItem",
					"dynamodb:UpdateItem",
					"dynamodb:DeleteItem"
				],
				"Resource": "%s"
			}]
		}`, table.Arn),
	})
	if err != nil {
		return nil, err
	}

	return &DeployOutput{
		TableName:      table.Name,
		TablePolicyArn: tablePolicy.Arn,
	}, nil
}
