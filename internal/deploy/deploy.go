// package deploy provides an operator used to deploy the zen infrastructure with pulumi.
package deploy

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/megakuul/zen/internal/deploy/email"
	"github.com/megakuul/zen/internal/deploy/leaderboard"
	"github.com/megakuul/zen/internal/deploy/manager"
	"github.com/megakuul/zen/internal/deploy/proxy"
	"github.com/megakuul/zen/internal/deploy/scheduler"
	"github.com/megakuul/zen/internal/deploy/storage"
	"github.com/megakuul/zen/internal/deploy/table"
	"github.com/megakuul/zen/internal/deploy/web"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Operator struct {
	region           string
	domains          []string
	autoDns          bool
	certificateArn   string
	deleteProtection bool
	buildCtxPath     string
	buildCachePath   string
}

type Option func(*Operator)

func New(region string, opts ...Option) *Operator {
	operator := &Operator{
		region:           region,
		autoDns:          true,
		domains:          []string{},
		certificateArn:   "",
		deleteProtection: false,
		buildCtxPath:     ".",
		buildCachePath:   "./.buildcache",
	}
	for _, opt := range opts {
		opt(operator)
	}
	return operator
}

// WithBuildPath defines a custom context and cache directory for the build step in the deployment.
// The context path is expected to contain the repository root (cmd/<handler>/<handler>.go).
func WithBuildPath(ctx, cache string) Option {
	return func(o *Operator) {
		o.buildCtxPath = ctx
		o.buildCachePath = cache
	}
}

// WithDomain adds the specified domain aliases to the public proxy endpoint.
// The first domain is used as primary endpoint and also as email domain, bounce.<first.domain> is used as envelope sender.
// If dnsSetup is enabled, an accessible hosted route53 zone for each domain is required.
func WithDomain(domains []string) Option {
	return func(o *Operator) {
		o.domains = domains
	}
}

// WithDnsSetup disables automatic dns, certificate and spf/dkim/dmarc management.
// Useful for externally hosted dns setups where your domains are not in route53.
// Provide a ready aws acm certificate from us-east-1 and be prepared to add certain domain entries manually in the process.
func WithDnsSetup(certArn string) Option {
	return func(o *Operator) {
		o.autoDns = false
		o.certificateArn = certArn
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
	leaderboardBuild, err := leaderboard.Build(ctx, &leaderboard.BuildInput{
		CtxPath:   o.buildCtxPath,
		CachePath: filepath.Join(o.buildCachePath, "lambda", "leaderboard"),
	})
	if err != nil {
		return fmt.Errorf("failed to build leaderboard function: %v", err)
	}
	managerBuild, err := manager.Build(ctx, &manager.BuildInput{
		CtxPath:   o.buildCtxPath,
		CachePath: filepath.Join(o.buildCachePath, "lambda", "manager"),
	})
	if err != nil {
		return fmt.Errorf("failed to build manager function: %v", err)
	}
	schedulerBuild, err := scheduler.Build(ctx, &scheduler.BuildInput{
		CtxPath:   o.buildCtxPath,
		CachePath: filepath.Join(o.buildCachePath, "lambda", "scheduler"),
	})
	if err != nil {
		return fmt.Errorf("failed to build scheduler function: %v", err)
	}
	webBuild, err := web.Build(ctx, &web.BuildInput{
		CtxPath:   o.buildCtxPath,
		CachePath: filepath.Join(o.buildCachePath, "web"),
	})
	if err != nil {
		return fmt.Errorf("failed to build web assets: %v", err)
	}
	tableDeploy, err := table.Deploy(ctx, &table.DeployInput{
		Region:           o.region,
		DeleteProtection: o.deleteProtection,
	})
	if err != nil {
		return fmt.Errorf("failed to deploy table: %v", err)
	}
	storageDeploy, err := storage.Deploy(ctx, &storage.DeployInput{
		Region:           o.region,
		DeleteProtection: o.deleteProtection,
		WebArtifacts:     webBuild.Artifacts,
	})
	if err != nil {
		return fmt.Errorf("failed to deploy storage: %v", err)
	}
	leaderboardDeploy, err := leaderboard.Deploy(ctx, &leaderboard.DeployInput{
		Region:          o.region,
		Handler:         leaderboardBuild.Handler,
		BucketPolicyArn: storageDeploy.BucketArn,
		BucketName:      storageDeploy.BucketName,
	})
	if err != nil {
		return fmt.Errorf("failed to deploy leaderboard system: %v", err)
	}
	schedulerDeploy, err := scheduler.Deploy(ctx, &scheduler.DeployInput{
		Region:         o.region,
		Handler:        schedulerBuild.Handler,
		TableName:      tableDeploy.TableName,
		TablePolicyArn: tableDeploy.TablePolicyArn,
		QueueName:      leaderboardDeploy.QueueName,
		QueuePolicyArn: leaderboardDeploy.QueuePolicyArn,
	})
	if err != nil {
		return fmt.Errorf("failed to deploy scheduler: %v", err)
	}
	emailDeploy, err := email.Deploy(ctx, &email.DeployInput{
		Region:  o.region,
		Domains: o.domains,
		AutoDns: o.autoDns,
	})
	if err != nil {
		return fmt.Errorf("failed to deploy ses email: %v", err)
	}
	_, err = manager.Deploy(ctx, &manager.DeployInput{
		Region:          o.region,
		Handler:         managerBuild.Handler,
		TableName:       tableDeploy.TableName,
		TablePolicyArn:  tableDeploy.TablePolicyArn,
		BucketName:      storageDeploy.BucketName,
		BucketPolicyArn: storageDeploy.BucketPolicyArn,
		EmailName:       emailDeploy.EmailName,
		EmailPolicyArn:  emailDeploy.EmailPolicyArn,
	})
	if err != nil {
		return fmt.Errorf("failed to deploy manager: %v", err)
	}
	proxyDeploy, err := proxy.Deploy(ctx, &proxy.DeployInput{
		Domains:        o.domains,
		AutoDns:        o.autoDns,
		CertificateArn: o.certificateArn,
		SchedulerDomain: schedulerDeploy.PublicUrl.ApplyT(func(input string) string {
			url, err := url.Parse(input)
			if err != nil {
				return "invalid.domain"
			}
			return url.Host
		}).(pulumi.StringOutput),
		ManagerDomain: schedulerDeploy.PublicUrl.ApplyT(func(input string) string {
			url, err := url.Parse(input)
			if err != nil {
				return "invalid.domain"
			}
			return url.Host
		}).(pulumi.StringOutput),
		BucketDomain: storageDeploy.BucketDomain,
	})
	if err != nil {
		return fmt.Errorf("failed to deploy proxy cdn: %v", err)
	}

	ctx.Export("ENDPOINT", proxyDeploy.ProxyDomain)
	return nil
}
