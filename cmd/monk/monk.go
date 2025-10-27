package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"github.com/megakuul/zen/cmd/monk/launch"
	"github.com/megakuul/zen/cmd/monk/nuke"
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
}

// appConfig is a json config used to store input parameters beside the pulumi state.
// Useful for updates where only certain parameters have to be changed.
// (pulumis IgnoreChange option is not sufficiently versatile for this use case).
type appConfig struct {
	Project          string   `json:"project"`
	Domains          []string `json:"domains"`
	AutoDns          bool     `json:"auto_dns"`
	DeleteProtection bool     `json:"delete_protection"`
	StateKey         string   `json:"state_key"`
}

func run(ctx context.Context) error {
	pterm.DefaultSpinner.Style = pterm.NewStyle(pterm.FgLightBlue)
	pterm.DefaultBasicText.Style = pterm.NewStyle(pterm.FgLightBlue)
	pterm.DefaultBasicText.Println("Welcome to the Zen monk bootstrapper ☯️")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	s3Client := s3.NewFromConfig(cfg)
	kmsClient := kms.NewFromConfig(cfg)

	bucket, prefix, err := setupBucket(ctx, s3Client)
	if err != nil {
		return fmt.Errorf("failed to setup bucket: %v", err)
	}
	configKey := fmt.Sprint(strings.TrimPrefix(prefix, "/"), "zen-app-config.json")

	var config appConfig
	configObject, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(configKey),
	})
	if err != nil {
		var nErr *s3types.NoSuchKey
		if !errors.As(err, &nErr) {
			return fmt.Errorf("failed to load app config: %v", err)
		}
		config.AutoDns = true
	} else {
		defer configObject.Body.Close()
		rawConfig, err := io.ReadAll(configObject.Body)
		if err != nil {
			return fmt.Errorf("failed to read app config: %v", err)
		}
		err = json.Unmarshal(rawConfig, &config)
		if err != nil {
			return fmt.Errorf("failed to parse app config: %v", err)
		}
	}

	if ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(config.Project == "").Show("Customize project name?"); ok {
		config.Project, _ = pterm.DefaultInteractiveTextInput.
			WithDefaultValue("zen").Show("Enter project name")
	}

	if ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(config.StateKey == "").Show("Customize state key?"); ok {
		config.StateKey, err = setupKey(ctx, kmsClient)
		if err != nil {
			return fmt.Errorf("failed to setup kms: %v", err)
		}
	}

	if ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(config.Domains == nil).Show("Customize domains?"); ok {
		config.Domains = []string{}
		for {
			domain, _ := pterm.DefaultInteractiveTextInput.
				WithDefaultValue("zen.megakuul.com").Show("Enter application domain")
			config.Domains = append(config.Domains, domain)
			ok, _ := pterm.DefaultInteractiveConfirm.
				WithDefaultValue(false).Show("Add another domain?")
			if !ok {
				break
			}
		}
	}

	if ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).Show("Customize delete protection?"); ok {
		config.DeleteProtection, _ = pterm.DefaultInteractiveConfirm.
			WithDefaultValue(false).Show("Enable delete protection?")
	}

	if ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).Show("Customize dns management?"); ok {
		config.AutoDns, _ = pterm.DefaultInteractiveConfirm.
			WithDefaultValue(true).Show("Enable dns management (requires route53 zone for each domain)?")
	}

	operatorOptions := []deploy.Option{}
	if config.DeleteProtection {
		operatorOptions = append(operatorOptions, deploy.WithDeleteProtection(true))
	}
	if !config.AutoDns {
		certArn, _ := pterm.DefaultInteractiveTextInput.
			WithDefaultValue("arn:aws:acm:us-east-1:...").Show("Enter acm certificate arn (must have SAN's for each domains)")
		operatorOptions = append(operatorOptions, deploy.WithDnsSetup(certArn))
	}
	operatorOptions = append(operatorOptions, deploy.WithBuildPath(".", ".buildcache"))
	operatorOptions = append(operatorOptions, deploy.WithDomain(config.Domains))
	operator := deploy.New(cfg.Region, operatorOptions...)

	ws, err := auto.NewLocalWorkspace(ctx, auto.Project(workspace.Project{
		Name:    tokens.PackageName(config.Project),
		Author:  aws.String("zen monk bootstrapper"),
		Runtime: workspace.NewProjectRuntimeInfo("go", map[string]any{}),
		Backend: &workspace.ProjectBackend{
			URL: fmt.Sprintf("s3://%s%s", bucket, prefix),
		},
	}),
		auto.SecretsProvider(fmt.Sprintf("awskms://%s", config.StateKey)),
		// TODO remove this workaround for an issue in pulumi <3.2.0
		// https://github.com/pulumi/pulumi/issues/7278
		auto.Stacks(map[string]workspace.ProjectStack{
			"prod": {SecretsProvider: fmt.Sprintf("awskms://%s", config.StateKey)},
			"test": {SecretsProvider: fmt.Sprintf("awskms://%s", config.StateKey)},
			"int":  {SecretsProvider: fmt.Sprintf("awskms://%s", config.StateKey)},
			"dev":  {SecretsProvider: fmt.Sprintf("awskms://%s", config.StateKey)},
		}),
		auto.Program(operator.Deploy),
	)
	if err != nil {
		return err
	}

	rawConfig, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize app config: %v", err)
	}
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(configKey),
		Body:   bytes.NewReader(rawConfig),
	})
	if err != nil {
		return fmt.Errorf("failed to upload app config: %v", err)
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
		return launch.Launch(ctx, ws)
	} else {
		action, _ := pterm.DefaultInteractiveSelect.
			WithOptions([]string{"launch", "nuke"}).Show("Select action")
		switch action {
		case "launch":
			return launch.Launch(ctx, ws)
		case "nuke":
			for _, stack := range stacks {
				err = nuke.Nuke(ctx, ws, stack.Name)
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

func setupKey(ctx context.Context, client *kms.Client) (string, error) {
	name, _ := pterm.DefaultInteractiveTextInput.
		WithDefaultValue("zen-state-key").Show("Enter key alias name")
	alias := fmt.Sprintf("alias/%s", name)

	spinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).
		Start("Creating kms key...")
	defer spinner.Stop()
	createResp, err := client.CreateKey(ctx, &kms.CreateKeyInput{
		KeySpec:     types.KeySpecSymmetricDefault,
		KeyUsage:    kmstypes.KeyUsageTypeEncryptDecrypt,
		Description: aws.String("Key used to encrypt sensitive pulumi stack data"),
	})
	if err != nil {
		return "", err
	}
	_, err = client.CreateAlias(ctx, &kms.CreateAliasInput{
		AliasName:   aws.String(alias),
		TargetKeyId: createResp.KeyMetadata.KeyId,
	})
	return alias, err
}

func setupBucket(ctx context.Context, client *s3.Client) (string, string, error) {
	ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(true).Show("Create new project?")
	if ok {
		name, _ := pterm.DefaultInteractiveTextInput.
			WithDefaultValue("zen-state-bucket").Show("Enter state bucket name")
		region, _ := pterm.DefaultInteractiveTextInput.
			WithDefaultValue("eu-central-1").Show("Enter state bucket name")
		spinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).
			Start("Creating state bucket...")
		defer spinner.Stop()
		_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(name),
			CreateBucketConfiguration: &s3types.CreateBucketConfiguration{
				LocationConstraint: s3types.BucketLocationConstraint(region),
			},
		})
		if err != nil {
			return "", "", err
		}
		return name, "/", nil
	} else {
		listResp, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
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
	}
}
