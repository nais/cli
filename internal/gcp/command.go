package gcp

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/nais/cli/internal/output"
)

func Login(ctx context.Context, out output.Output) error {
	return executeCommand(ctx, out, "auth", "login", "--update-adc")
}

func Logout(ctx context.Context, out output.Output) error {
	if err := executeCommand(ctx, out, "auth", "application-default", "revoke", "--quiet"); err != nil {
		return err
	}

	return executeCommand(ctx, out, "auth", "revoke")
}

func executeCommand(ctx context.Context, out output.Output, arg ...string) error {
	cmd := exec.CommandContext(ctx, "gcloud", arg...)
	o, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\nerror running %q command: %w", string(o), cmd.String(), err)
	}

	out.Println(string(o))

	return nil
}
