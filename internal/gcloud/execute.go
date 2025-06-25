// Package gcloud is a wrapper for gcloud commands
package gcloud

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/nais/cli/pkg/cli"
)

func executeGcloud(ctx context.Context, out cli.Output, arg ...string) error {
	cmd := exec.CommandContext(ctx, "gcloud", arg...)
	o, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\nerror running %q command: %w", string(o), cmd.String(), err)
	}

	out.Println(string(o))

	return nil
}
