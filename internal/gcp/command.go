package gcp

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/nais/cli/internal/root"
)

func Login(ctx context.Context, _ *root.Flags) error {
	return executeCommand(ctx, "auth", "login", "--update-adc")
}

func Logout(ctx context.Context, _ *root.Flags) error {
	if err := executeCommand(ctx, "auth", "application-default", "revoke", "--quiet"); err != nil {
		return err
	}

	return executeCommand(ctx, "auth", "revoke")
}

func executeCommand(ctx context.Context, arg ...string) error {
	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "gcloud", arg...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v\nerror running %q command: %w", buf.String(), cmd.String(), err)
	}

	return nil
}
