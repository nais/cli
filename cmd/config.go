package cmd

import (
	"fmt"
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
