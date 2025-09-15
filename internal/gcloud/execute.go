// Package gcloud is a wrapper for gcloud commands
package gcloud

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/nais/naistrix"
)

func executeGcloud(ctx context.Context, out naistrix.Output, verbose bool, arg ...string) error {
	cmd := exec.CommandContext(ctx, "gcloud", arg...)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("running: %q, err %w", cmd.String(), err)
		}
	} else {
		o, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%v\nerror running %q command: %w", string(o), cmd.String(), err)
		}
		out.Println("Logged in with gcloud.")
	}

	return nil
}
