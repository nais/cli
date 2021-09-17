package helpers

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"runtime"
	"strings"
)

const (
	FilePermission = 0775
)

func DefaultDestination(dest string) (string, error) {
	current, err := os.Getwd()
	if dest == "" {
		return current, nil
	}

	if err != nil {
		return "", fmt.Errorf("could assign directory; %s", err)
	}

	dest = system(dest)
	newPath := fmt.Sprintf("%s%s", current, dest)
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		if err = os.Mkdir(newPath, os.FileMode(FilePermission)); err != nil {
			return "", fmt.Errorf("could not create directory; %s", err)
		}
	}
	return newPath, nil
}

func system(dest string) string {
	if runtime.GOOS == "windows" {
		if !strings.HasPrefix(dest, "\\") {
			return fmt.Sprintf("\\%s", dest)
		} else {
			return dest
		}
	}
	if !strings.HasPrefix(dest, "/") {
		return fmt.Sprintf("/%s", dest)
	}
	return dest
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
