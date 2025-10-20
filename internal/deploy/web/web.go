package web

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type BuildInput struct {
	CtxPath   string
	CachePath string
}

type BuildOutput struct {
	Artifacts pulumi.AssetOrArchiveMapOutput
}

func Build(ctx *pulumi.Context, input *BuildInput) (*BuildOutput, error) {
	outputPath, err := filepath.Abs(input.CachePath)
	if err != nil {
		return nil, err
	}
	command := fmt.Sprintf("npm -C ./web run build")
	build, err := local.NewCommand(ctx, "web", &local.CommandArgs{
		Create:     pulumi.String(command),
		Update:     pulumi.String(command),
		Dir:        pulumi.String(input.CtxPath),
		AssetPaths: pulumi.ToStringArray([]string{fmt.Sprintf("%s/**", outputPath)}),
		Environment: pulumi.ToStringMap(map[string]string{
			"BUILD_DIR": outputPath,
		}),
		Logging: local.LoggingStdoutAndStderr,
		// not rebuilding causes the empty archive to trigger a replacement of the current webassets.
		// therefore, rebuild is always triggered.
		Triggers: pulumi.ToArray([]any{uuid.New().String()}),
	})
	if err != nil {
		return nil, err
	}
	return &BuildOutput{
		Artifacts: build.Assets,
	}, nil
}
