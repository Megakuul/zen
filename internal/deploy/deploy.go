// package deploy provides an operator used to deploy the zen infrastructure with pulumi.
package deploy

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Operator struct {
	logger           *slog.Logger
	region           string
	domains          []string
	autoDns          bool
	certificateArn   string
	deleteProtection bool
	buildCtxPath     string
	buildCachePath   string
}

type Option func(*Operator)

func New(logger *slog.Logger, region string, opts ...Option) *Operator {
	operator := &Operator{
		logger:           logger,
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
	handlerPath, err := filepath.Abs(filepath.Join(o.buildCachePath, "lambda", "leaderboard"))
	if err != nil {
		return err
	}
	command := fmt.Sprintf("go build -o '%s' ./cmd/leaderboard/leaderboard.go", handlerPath)
	build, err := local.NewCommand(ctx, "leaderboard", &local.CommandArgs{
		Create:       pulumi.String(command),
		Update:       pulumi.String(command),
		Dir:          pulumi.String(o.buildCtxPath),
		ArchivePaths: pulumi.ToStringArray([]string{handlerPath}),
		Environment: pulumi.ToStringMap(map[string]string{
			"CGO_ENABLED": "0",
			"GOOS":        "linux",
			"GOARCH":      "arm64",
		}),
		Logging: local.LoggingStdoutAndStderr,
	})
	if err!=nil {
		return nil, err
	}
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
		BucketName:      storageOutputs.BucketName,
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
	emailOutputs, err := o.deployEmail(ctx, &emailInput{})
	if err != nil {
		return fmt.Errorf("failed to deploy ses email: %v", err)
	}
	_, err = o.deployManager(ctx, &managerInput{
		CodeArchive:     pulumi.NewAssetArchive(map[string]any{}),
		TableName:       tableOutputs.TableName,
		TablePolicyArn:  tableOutputs.TablePolicyArn,
		BucketName:      storageOutputs.BucketName,
		BucketPolicyArn: storageOutputs.BucketPolicyArn,
		EmailName:       emailOutputs.EmailName,
		EmailPolicyArn:  emailOutputs.EmailPolicyArn,
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

	ctx.Export("ENDPOINT", proxyOutput.ProxyDomain)
	return nil
}

// lookupZone checks if there is a route53 zone for the provided domain.
// It traverses each domain segment to check for a zone.
func lookupZone(ctx *pulumi.Context, domain string) (*route53.LookupZoneResult, error) {
	var (
		err      error
		zoneName string = domain
		zone     *route53.LookupZoneResult
	)
	for {
		if lZone, lErr := route53.LookupZone(ctx, &route53.LookupZoneArgs{
			Name:        pulumi.StringRef(zoneName),
			PrivateZone: pulumi.BoolRef(false),
		}); lErr != nil {
			err = errors.Join(err, lErr)
		} else {
			zone = lZone
			break
		}
		segments := strings.Split(zoneName, ".")
		if len(segments) < 3 { // 3 segments minimum, the tld is never a hosted zone
			return nil, fmt.Errorf("no route53 hosted zone found for domain '%s': %v", domain, err)
		}
		zoneName = segments[1]
	}
	return zone, nil
}
