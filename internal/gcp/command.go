package gcp

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/nais/cli/internal/output"
)

func Login(ctx context.Context, output output.Output) error {
	return executeCommand(ctx, output, "auth", "login", "--update-adc")
}

func Logout(ctx context.Context, output output.Output) error {
	if err := executeCommand(ctx, output, "auth", "application-default", "revoke", "--quiet"); err != nil {
		return err
	}

	return executeCommand(ctx, output, "auth", "revoke")
}

func executeCommand(ctx context.Context, output output.Output, arg ...string) error {
	cmd := exec.CommandContext(ctx, "gcloud", arg...)
	o, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\nerror running %q command: %w", string(o), cmd.String(), err)
	}

	output.Println(string(o))

	return nil
}
