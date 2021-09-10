package helpers

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

const (
	FilePermission = 0775
)

func DefaultDestination(dest string) (string, error) {
	path, err := os.Getwd()
	if dest == "" {
		return path, nil
	}

	if !strings.HasPrefix(dest, "/") {
		dest = fmt.Sprintf("/%s", dest)
	}

	newPath := fmt.Sprintf("%s%s", path, dest)
	if err != nil {
		return "", fmt.Errorf("could assign directory; %s", err)
	}

	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		if err = os.Mkdir(newPath, os.FileMode(FilePermission)); err != nil {
			return "", fmt.Errorf("could not create directory; %s", err)
		}
	}
	return newPath, nil
}

func GetString(cmd *cobra.Command, flag string, required bool) (string, error) {
	env := viper.GetString(flag)
	if env != "" {
		return env, nil
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

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
