package nuke

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
)

// Nuke performs an interactive process to destroy a running operator.
func Nuke(ctx context.Context, ws auto.Workspace, stackName string) error {
	stack, err := auto.SelectStack(ctx, stackName, ws)
	if err != nil {
		return fmt.Errorf("failed to load stack: %v", err)
	}
	spinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).
		Start("Loading destruction preview...")
	defer spinner.Stop()
	preview, err := stack.PreviewDestroy(ctx, optdestroy.Color("always"))
	if err != nil {
		return fmt.Errorf("stack dry run failed: %v", err)
	}
	spinner.Stop()
	fmt.Println()
	fmt.Println(preview.StdOut)
	if preview.StdErr != "" {
		fmt.Println()
		fmt.Println("⚠️ Anomalies detected in destruction preview")
	}
	fmt.Println()
	ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).Show("Destroy the stack?")
	if !ok {
		return fmt.Errorf("process cancelled")
	}
	multi, _ := pterm.DefaultMultiPrinter.Start()
	defer multi.Stop()
	stackWriter := multi.NewWriter()
	spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(true).
		WithWriter(multi.NewWriter()).Start("Destroying application...")
	_, err = stack.Destroy(ctx, optdestroy.ProgressStreams(stackWriter))
	if err != nil {
		return fmt.Errorf("failed to destroy stack: %v", err)
	}
	pterm.DefaultBasicText.Println("Success 🎉 Your application got wiped from earth ☢️")
	return nil
}
