package aiven

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nais/cli/pkg/cli"
)

func TidyLocalSecrets(out cli.Output) error {
	aivenSecretFolders, err := findFoldersToRemove()
	if err != nil {
		return err
	}

	return tidy(aivenSecretFolders, out)
}

func tidy(folders []string, out cli.Output) error {
	if len(folders) > 0 {
		for _, folder := range folders {
			out.Printf("Deleting: %s\n", folder)
			err := os.RemoveAll(folder)
			if err != nil {
				return fmt.Errorf("failed deleting %v: %v", folder, err)
			}
		}
	} else {
		out.Println("All tidy")
	}
	return nil
}

func findFoldersToRemove() ([]string, error) {
	var folders []string
	err := filepath.Walk(os.TempDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() && strings.Contains(path, FolderPrefix) {
			folders = append(folders, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking path: %w", err)
	}

	return folders, nil
}
