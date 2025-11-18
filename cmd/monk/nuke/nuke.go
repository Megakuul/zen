package nuke

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
)

// Nuke performs an interactive process to destroy a running application.
func Nuke(ctx context.Context, ws auto.Workspace, stackName string) error {
	stack, err := auto.SelectStack(ctx, stackName, ws)
	if err != nil {
		return fmt.Errorf("failed to load stack: %v", err)
	}
	ok, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).Show(fmt.Sprintf("Destroy the stack (%s)?", stackName))
	if !ok {
		return fmt.Errorf("process cancelled")
	}
	multi, _ := pterm.DefaultMultiPrinter.Start()
	defer multi.Stop()
	stackWriter := multi.NewWriter()
	spinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).
		WithWriter(multi.NewWriter()).Start(fmt.Sprintf("Destroying Zen system (%s)...", stackName))
	defer spinner.Stop()
	_, err = stack.Destroy(ctx, optdestroy.ProgressStreams(stackWriter))
	if err != nil {
		return fmt.Errorf("failed to destroy stack '%s': %v", stackName, err)
	}
	pterm.DefaultBasicText.Println("Success 🎉 Zen got wiped from earth ☢️")
	return nil
}
