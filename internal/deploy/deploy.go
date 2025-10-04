// package deploy provides an operator used to deploy the zen infrastructure with pulumi.
package deploy

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Operator struct {
	logger           *slog.Logger
	region           string
	domains          []string
	hostedZone       string
	certificateArn   string
	deleteProtection bool
}

type Option func(*Operator)

func New(logger *slog.Logger, region string, opts ...Option) *Operator {
	operator := &Operator{
		logger:           logger,
		region:           region,
		domains:          []string{},
		certificateArn:   "",
		deleteProtection: false,
	}
	for _, opt := range opts {
		opt(operator)
	}
	return operator
}

// WithDomain adds the specified domain aliases to the public proxy endpoint.
// If no explicit certificate arn is provided, an accessible hosted route53 zone for each domain is required.
func WithDomain(domains []string) Option {
	return func(o *Operator) {
		o.domains = domains
	}
}

// WithCertificate uses an existing acm certificate instead of creating one via hosted zone.
// The certificate must be located in us-east-1. Useful for externally hosted dns zones.
// If this option is enabled, dns records to the proxy must be added manually.
func WithCertificate(arn string) Option {
	return func(o *Operator) {
		o.certificateArn = arn
	}
}

// WithDeleteProtection enables delete protection mechanisms
// for database tables and storage (useful for production environments).
func WithDeleteProtection(enable bool) Option {
	return func(o *Operator) {
		o.deleteProtection = enable
	}
}

func (o *Operator) Deploy(ctx *pulumi.Context) error {
	tableOutputs, err := o.deployTable(ctx, &tableInput{})
	if err != nil {
		return fmt.Errorf("failed to deploy table: %v", err)
	}
	storageOutputs, err := o.deployStorage(ctx, &storageInput{})
	if err != nil {
		return fmt.Errorf("failed to deploy storage: %v", err)
	}
	leaderboardOutputs, err := o.deployLeaderboard(ctx, &leaderboardInput{
		CodeArchive:     pulumi.NewAssetArchive(map[string]any{}),
		BucketPolicyArn: storageOutputs.BucketArn,
	})
	if err != nil {
		return fmt.Errorf("failed to deploy leaderboard system: %v", err)
	}
	schedulerOutputs, err := o.deployScheduler(ctx, &schedulerInput{
		CodeArchive:    pulumi.NewAssetArchive(map[string]any{}),
		TableName:      tableOutputs.TableName,
		TablePolicyArn: tableOutputs.TablePolicyArn,
		QueueName:      leaderboardOutputs.QueueName,
		QueuePolicyArn: leaderboardOutputs.QueuePolicyArn,
	})
	if err != nil {
		return fmt.Errorf("failed to deploy scheduler: %v", err)
	}
	_, err = o.deployManager(ctx, &managerInput{
		CodeArchive: pulumi.NewAssetArchive(map[string]any{}),
		TableName:   tableOutputs.TableName,
	})
	if err != nil {
		return fmt.Errorf("failed to deploy manager: %v", err)
	}
	proxyOutput, err := o.deployProxy(ctx, &proxyInput{
		SchedulerDomain: schedulerOutputs.PublicUrl.ApplyT(func(input string) string {
			url, err := url.Parse(input)
			if err != nil {
				return "invalid.domain"
			}
			return url.Host
		}).(pulumi.StringOutput),
		ManagerDomain: schedulerOutputs.PublicUrl.ApplyT(func(input string) string {
			url, err := url.Parse(input)
			if err != nil {
				return "invalid.domain"
			}
			return url.Host
		}).(pulumi.StringOutput),
		BucketDomain: storageOutputs.BucketDomain,
	})
	if err != nil {
		return fmt.Errorf("failed to deploy proxy cdn: %v", err)
	}

	ctx.Export("endpoint", proxyOutput.ProxyDomain)
	return nil
}
