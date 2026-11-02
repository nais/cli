package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestMainWritesErrorsToStderr(t *testing.T) {
	if os.Getenv("NAIS_TEST_MAIN_ERROR") == "1" {
		os.Args = []string{"nais", "unknown-command-for-test"}
		main()
		return
	}

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, "-test.run=^TestMainWritesErrorsToStderr$")
	cmd.Env = append(os.Environ(), "NAIS_TEST_MAIN_ERROR=1", "DO_NOT_TRACK=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if exit, ok := errors.AsType[*exec.ExitError](cmd.Run()); !ok || exit.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got %v; stderr: %s", exit, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unknown command "unknown-command-for-test"`) {
		t.Errorf("expected command error on stderr, got %q", stderr.String())
	}
}
