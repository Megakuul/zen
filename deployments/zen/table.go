package zen

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/dynamodb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type tableInput struct { }

type tableOutput struct {
	TablePolicyArn pulumi.StringOutput
}

func (o *Operator) deployTable(ctx *pulumi.Context, input *tableInput) (*tableOutput, error) {
	table, err := dynamodb.NewTable(ctx, "table", &dynamodb.TableArgs{
		Name: pulumi.String("zen-table"),
		Region: pulumi.String(o.region),
		BillingMode: pulumi.String("PAY_PER_REQUEST"),
		HashKey: pulumi.String("pk"),
		RangeKey: pulumi.String("sk"),
		Attributes: dynamodb.TableAttributeArray{
			dynamodb.TableAttributeArgs{Name: pulumi.String("pk"), Type: pulumi.String("S")},
			dynamodb.TableAttributeArgs{Name: pulumi.String("sk"), Type: pulumi.String("S")},
		},
		OnDemandThroughput: dynamodb.TableOnDemandThroughputPtr(&dynamodb.TableOnDemandThroughputArgs{
			MaxWriteRequestUnits: pulumi.IntPtr(10),
			MaxReadRequestUnits: pulumi.IntPtr(100),
		}),
		Ttl: dynamodb.TableTtlPtr(&dynamodb.TableTtlArgs{
			AttributeName: pulumi.String("expiry"),
			Enabled: pulumi.BoolPtr(true),
		}),
		DeletionProtectionEnabled: pulumi.BoolPtr(o.deleteProtection),
	})
	if err!=nil {
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
					"dynamodb:UpdateItem"
				],
				"Resource": "%s"
			}]
		}`, table.Arn),
	})
	if err != nil {
		return nil, err
	}

	return &tableOutput{
		TablePolicyArn: tablePolicy.Arn,
	}, nil
}
