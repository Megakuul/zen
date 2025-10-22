package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/megakuul/zen/internal/deploy"
	"github.com/pterm/pterm"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigs:
			cancel()
			os.Exit(1)
		case <-ctx.Done():
			return
		}
	}()

	if err := run(ctx); err != nil {
		pterm.DefaultBasicText.Println("\n❌ ========= ERROR =========")
		pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.FgRed)).Println(strings.TrimSpace(err.Error()))
		pterm.DefaultBasicText.Println("❌ ========= ERROR =========")
		os.Exit(1)
	}
	return
}

func run(ctx context.Context) error {
	pterm.DefaultSpinner.Style = pterm.NewStyle(pterm.FgLightBlue)
	pterm.DefaultBasicText.Style = pterm.NewStyle(pterm.FgLightBlue)
	pterm.DefaultBasicText.Println("Welcome to the Zen monk bootstrapper ☯️")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	project, _ := pterm.DefaultInteractiveTextInput.
		WithDefaultValue("zen").Show("Enter project name")

	operatorOptions := []deploy.Option{}

	ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).Show("Enable delete protection?")
	if ok {
		operatorOptions = append(operatorOptions, deploy.WithDeleteProtection(true))
	}

	domains := []string{}
	for {
		domain, _ := pterm.DefaultInteractiveTextInput.
			WithDefaultValue("zen.megakuul.com").Show("Enter application domain")
		domains = append(domains, domain)
		ok, _ := pterm.DefaultInteractiveConfirm.
			WithDefaultValue(false).Show("Add another domain?")
		if !ok {
			break
		}
	}

	ok, _ = pterm.DefaultInteractiveConfirm.
		WithDefaultValue(true).Show("Enable dns management (requires route53 zone for each domain)?")
	if !ok {
		certArn, _ := pterm.DefaultInteractiveTextInput.
			WithDefaultValue("arn:aws:acm:us-east-1:...").Show("Enter acm certificate arn (must have SAN's for each domains)")
		operatorOptions = append(operatorOptions, deploy.WithDnsSetup(certArn))
	}

	operatorOptions = append(operatorOptions, deploy.WithBuildPath(".", ".buildcache"))
	operatorOptions = append(operatorOptions, deploy.WithDomain(domains))
	operator := deploy.New(cfg.Region, operatorOptions...)

	bucket, prefix, err := setupBucket(ctx, cfg, project)
	if err != nil {
		return fmt.Errorf("failed to setup bucket: %v", err)
	}

	keyAlias, err := setupKey(ctx, cfg, project)
	if err != nil {
		return fmt.Errorf("failed to setup kms: %v", err)
	}

	ws, err := auto.NewLocalWorkspace(ctx, auto.Project(workspace.Project{
		Name:    tokens.PackageName(project),
		Author:  aws.String("zen monk bootstrapper"),
		Runtime: workspace.NewProjectRuntimeInfo("go", map[string]any{}),
		Backend: &workspace.ProjectBackend{
			URL: fmt.Sprintf("s3://%s%s", bucket, prefix),
		},
	}),
		auto.SecretsProvider(fmt.Sprintf("awskms://%s", keyAlias)),
		auto.Program(operator.Deploy),
	)
	if err != nil {
		return err
	}

	spinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).
		Start("Searching for existing stacks...")
	defer spinner.Stop()
	stacks, err := ws.ListStacks(ctx)
	if err != nil {
		return err
	}
	spinner.Stop()
	if len(stacks) < 1 {
		return launch(ctx, ws)
	} else {
		action, _ := pterm.DefaultInteractiveSelect.
			WithOptions([]string{"launch", "nuke"}).Show("Select action")
		switch action {
		case "launch":
			return launch(ctx, ws)
		case "nuke":
			for _, stack := range stacks {
				err = nuke(ctx, ws, stack.Name)
				if err != nil {
					return err
				}
			}
			return nil
		default:
			return fmt.Errorf("not a valid action")
		}
	}
}

func setupKey(ctx context.Context, cfg aws.Config, project string) (string, error) {
	kmsClient := kms.NewFromConfig(cfg)
	ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).Show("Use existing kms key for state encryption?")
	if ok {
		listResp, err := kmsClient.ListAliases(ctx, &kms.ListAliasesInput{})
		if err != nil {
			return "", err
		}
		keys := []string{}
		for _, alias := range listResp.Aliases {
			keys = append(keys, *alias.AliasName)
		}
		selected, err := pterm.DefaultInteractiveSelect.
			WithOptions(keys).
			Show("Select kms key")
		if err != nil {
			return "", err
		}
		return selected, nil
	} else {
		name, _ := pterm.DefaultInteractiveTextInput.
			WithDefaultValue(fmt.Sprint(project, "-state-key")).Show("Enter key alias name")
		alias := fmt.Sprintf("alias/%s", name)

		spinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).
			Start("Creating kms key...")
		defer spinner.Stop()
		createResp, err := kmsClient.CreateKey(ctx, &kms.CreateKeyInput{
			KeySpec:     types.KeySpecSymmetricDefault,
			KeyUsage:    kmstypes.KeyUsageTypeEncryptDecrypt,
			Description: aws.String("Key used to encrypt sensitive pulumi stack data"),
		})
		if err != nil {
			return "", err
		}
		_, err = kmsClient.CreateAlias(ctx, &kms.CreateAliasInput{
			AliasName:   aws.String(alias),
			TargetKeyId: createResp.KeyMetadata.KeyId,
		})
		return alias, err
	}
}

func setupBucket(ctx context.Context, cfg aws.Config, project string) (string, string, error) {
	s3Client := s3.NewFromConfig(cfg)
	ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).Show("Use existing s3 bucket for state?")
	if ok {
		listResp, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
		if err != nil {
			return "", "", err
		}
		buckets := []string{}
		for _, bucket := range listResp.Buckets {
			buckets = append(buckets, *bucket.Name)
		}
		selected, err := pterm.DefaultInteractiveSelect.
			WithOptions(buckets).
			Show("Select bucket")
		if err != nil {
			return "", "", err
		}
		prefix, _ := pterm.DefaultInteractiveTextInput.WithDefaultValue("/").Show("Specify bucket prefix")
		return selected, prefix, nil
	} else {
		name, _ := pterm.DefaultInteractiveTextInput.
			WithDefaultValue(fmt.Sprint(project, "-state-bucket")).Show("Enter state bucket name")
		region, _ := pterm.DefaultInteractiveTextInput.
			WithDefaultValue("eu-central-1").Show("Enter state bucket name")
		spinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).
			Start("Creating state bucket...")
		defer spinner.Stop()
		_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(name),
			CreateBucketConfiguration: &s3types.CreateBucketConfiguration{
				LocationConstraint: s3types.BucketLocationConstraint(region),
			},
		})
		if err != nil {
			return "", "", err
		}
		return name, "/", nil
	}
}
