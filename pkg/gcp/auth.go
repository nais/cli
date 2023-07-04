package gcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func ValidateUserLogin(ctx context.Context, enforceNais bool) error {
	args := []string{
		"config",
		"list",
		"account",
		"--format", "value(core.account)",
	}

	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%v\nerror running '%v' command: %w", buf.String(), cmd.String(), err)
	}

	user := strings.TrimSpace(buf.String())
	if user == "" {
		return fmt.Errorf("missing active user, have you logged in with 'gcloud auth login --update-adc'")
	}

	if enforceNais && !strings.HasSuffix(user, "@nais.io") {
		return fmt.Errorf("active gcloud-user is not a nais.io-user: %v", user)
	}

	_, exists := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if exists {
		return nil
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := homedir + "/.config/gcloud/application_default_credentials.json"

	if runtime.GOOS == "windows" {
		configDir = "%APPDATA%\\gcloud\\application_default_credentials.json"
	}

	_, err = os.Stat(configDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("you are missing Application Default Credentials, run `gcloud auth application-default login` first")
		}
		return err
	}

	return nil
}

func GetActiveUserEmail(ctx context.Context) (string, error) {
	args := []string{
		"config",
		"list",
		"account",
		"--format", "value(core.account)",
	}

	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%v\nerror running '%v' command: %w", buf.String(), cmd.String(), err)
	}

	user := strings.TrimSpace(buf.String())
	if user == "" {
		return "", fmt.Errorf("no users found, are you logged in")
	}

	return user, nil
}
