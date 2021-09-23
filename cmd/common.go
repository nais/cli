package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

const (
	AivenSecretFolderPrefix = "aiven-secret-"
)

func DefaultDestination(dest string) (string, error) {
	if dest != "" {
		newPath, err := filepath.Abs(dest)
		if err != nil {
			return "", fmt.Errorf("unable to make %s absolute: %w", dest, err)
		}
		return newPath, nil
	}

	newPath, err := os.MkdirTemp("", AivenSecretFolderPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	return newPath, nil
}

func GetString(cmd *cobra.Command, flag string, required bool) (string, error) {
	if viper.GetString(flag) != "" {
		return viper.GetString(flag), nil
	}
	arg, err := cmd.Flags().GetString(flag)
	if err != nil {
		return "", fmt.Errorf("getting %s: %s", flag, err)
	}
	if arg == "" {
		if required {
			return "", fmt.Errorf("%s is reqired", flag)
		}
	}
	return arg, nil
}
