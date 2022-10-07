package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	AivenSecretFolderPrefix = "aiven-secret-"
)

func DefaultDestination() (string, error) {
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
		return "", fmt.Errorf("flag '--%s': %s", flag, err)
	}
	if arg == "" {
		if required {
			return "", fmt.Errorf("%s is required", flag)
		}
	}
	return arg, nil
}

func GetInt(cmd *cobra.Command, flag string, required bool) (int, error) {
	if viper.GetInt(flag) != 0 {
		return viper.GetInt(flag), nil
	}
	arg, err := cmd.Flags().GetInt(flag)
	if err != nil {
		return 0, fmt.Errorf("getting '--%s': %s", flag, err)
	}
	if arg == 0 {
		if required {
			return 0, fmt.Errorf("%s is required", flag)
		}
	}
	return arg, nil
}
