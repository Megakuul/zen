package web

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type BuildInput struct {
	CtxPath string
}

type BuildOutput struct {
	Artifacts pulumi.AssetOrArchiveMapOutput
	Router    pulumi.String
}

func Build(ctx *pulumi.Context, input *BuildInput) (*BuildOutput, error) {
	contextPath, err := filepath.Abs(input.CtxPath)
	if err != nil {
		return nil, err
	}
	commandPath := filepath.Join(contextPath, ".cache/web")
	if err := os.MkdirAll(commandPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache path: %v", err)
	}
	command := "npm -C ../../web run build"
	build, err := local.NewCommand(ctx, "web", &local.CommandArgs{
		Create: pulumi.String(command),
		Update: pulumi.String(command),
		// must be inside cache otherwise the output archive contains cache paths
		Dir:        pulumi.String(commandPath),
		AssetPaths: pulumi.ToStringArray([]string{"**"}),
		Environment: pulumi.ToStringMap(map[string]string{
			"BUILD_DIR": commandPath,
		}),
		Logging: local.LoggingStderr,
		// not rebuilding causes the empty archive to trigger a replacement of the current webassets.
		// therefore, rebuild is always triggered.
		Triggers: pulumi.ToArray([]any{uuid.New().String()}),
	})
	if err != nil {
		return nil, err
	}
	router, err := os.ReadFile(filepath.Join(input.CtxPath, "web/router/router.js"))
	if err != nil {
		return nil, fmt.Errorf("no web router found: %v", err)
	}
	return &BuildOutput{
		Artifacts: build.Assets,
		Router:    pulumi.String(router),
	}, nil
}
