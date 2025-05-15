package aiven

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func TidyLocalSecrets() error {
	aivenSecretFolders, err := findFoldersToRemove()
	if err != nil {
		return err
	}

	return tidy(aivenSecretFolders)
}

func tidy(folders []string) error {
	if len(folders) > 0 {
		for _, folder := range folders {
			fmt.Printf("Deleting: %s\n", folder)
			err := os.RemoveAll(folder)
			if err != nil {
				return fmt.Errorf("failed deleting %v: %v", folder, err)
			}
		}
	} else {
		fmt.Println("All tidy")
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
