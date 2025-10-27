package launch

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
)

// Launch performs an interactive process to deploy the operator stack on the provided workspace.
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
	spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(true).
		Start("Loading stack update preview...")
	preview, err := stack.Preview(ctx, optpreview.Color("always"))
	if err != nil {
		return fmt.Errorf("stack preview failed: %v", err)
	}
	spinner.Stop()
	fmt.Println()
	fmt.Println(preview.StdOut)
	if preview.StdErr != "" {
		fmt.Println()
		fmt.Println("⚠️ Anomalies detected in deployment preview")
	}
	fmt.Println()
	ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).Show("Deploy the operator?")
	if !ok {
		return fmt.Errorf("process cancelled")
	}
	multi, _ := pterm.DefaultMultiPrinter.Start()
	defer multi.Stop()
	stackWriter := multi.NewWriter()
	spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(true).
		WithWriter(multi.NewWriter()).Start("Building and deploying application...")
	err = stack.SetAllConfig(ctx, auto.ConfigMap{
		"aws:defaultTags": auto.ConfigValue{Value: fmt.Sprintf(`{
			"tags": {
				"system": "zen",
				"environment": "%s"
			}
		}`, environment)},
	})
	if err!=nil {
		return fmt.Errorf("failed to set default tags: %v", err)
	}
	result, err := stack.Up(ctx, optup.ProgressStreams(stackWriter))
	if err != nil {
		return fmt.Errorf("failed to update stack: %v", err)
	}
	pterm.DefaultBasicText.Println("Success 🎉 Use the outputs to finish setup ⛩️")
	for k, v := range result.Outputs {
		pterm.DefaultBasicText.Printf(" - %s: %v\n", k, v.Value)
	}

	return nil
}
