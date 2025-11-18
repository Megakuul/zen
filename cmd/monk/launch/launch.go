package launch

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
)

// Launch performs an interactive process to deploy the stack on the provided workspace.
func Launch(ctx context.Context, ws auto.Workspace) error {
	environment, _ := pterm.DefaultInteractiveSelect.WithOptions([]string{
		"prod", "test", "int", "dev",
	}).WithDefaultOption("prod").Show("Select environment")
	spinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).
		Start("Merging stack update locally...")
	defer spinner.Stop()
	stack, err := auto.UpsertStack(ctx, environment, ws)
	if err != nil {
		return fmt.Errorf("failed to construct stack: %v", err)
	}
	spinner.Stop()
	ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).Show("Deploy the Zen system?")
	if !ok {
		return fmt.Errorf("process cancelled")
	}
	refresh, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).Show("Enable refresh mode (fix state drifts)?")

	multi, _ := pterm.DefaultMultiPrinter.Start()
	stackWriter := multi.NewWriter()
	spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(true).
		WithWriter(multi.NewWriter()).Start("Building and deploying Zen system...")
	err = stack.SetAllConfig(ctx, auto.ConfigMap{
		"aws:defaultTags": auto.ConfigValue{Value: fmt.Sprintf(`{
			"tags": {
				"system": "zen",
				"environment": "%s"
			}
		}`, environment)},
	})
	if err != nil {
		multi.Stop()
		return fmt.Errorf("failed to set default tags: %v", err)
	}

	opts := []optup.Option{optup.ProgressStreams(stackWriter)}
	if refresh {
		opts = append(opts, optup.Refresh())
	}
	result, err := stack.Up(ctx, opts...)
	if err != nil {
		multi.Stop()
		return fmt.Errorf("failed to update stack: %v", err)
	}
	multi.Stop()
	pterm.DefaultBasicText.Println("Success 🎉 Use the outputs to finish setup ⛩️")
	for k, v := range result.Outputs {
		pterm.DefaultBasicText.Printf(" - %s: %v\n", k, v.Value)
	}

	return nil
}
