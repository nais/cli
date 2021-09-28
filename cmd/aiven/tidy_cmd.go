package aiven

import (
	"fmt"
	"github.com/nais/nais-cli/cmd"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type AivenSecretFolder struct {
	Abs string
}

var TidyCmd = &cobra.Command{
	Use:     "tidy",
	Short:   "Remove '/tmp' folder '$TMPDIR' and files created by the aiven command",
	Long:    "Caution - This will delete all files in '/tmp' folder starting with 'aiven-secret-'",
	Example: `nais aiven tidy`,
	RunE: func(command *cobra.Command, args []string) error {

		aivenSecretFolders, err := findFoldersToTidy()
		if err != nil {
			return fmt.Errorf("walking folders")
		}

		err = Tidy(aivenSecretFolders)
		if err != nil {
			return err
		}
		return nil
	},
}

func Tidy(aivenSecretFolders []AivenSecretFolder) error {
	if len(aivenSecretFolders) > 0 {
		for _, folder := range aivenSecretFolders {
			log.Default().Printf("tidy: %s", folder.Abs)
			err := os.RemoveAll(folder.Abs)
			if err != nil {
				return fmt.Errorf("tidying folder: %s", folder.Abs)
			}
		}
	} else {
		log.Default().Println("all tidy")
	}
	return nil
}

func findFoldersToTidy() ([]AivenSecretFolder, error) {
	var aivenSecretFolders []AivenSecretFolder
	err := filepath.Walk(os.TempDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Walking folder we ignore folders we can not operate on.
			return nil
		}
		// Keep AivenSecretfolders we want to delete later
		if info.IsDir() && strings.Contains(path, cmd.AivenSecretFolderPrefix) {
			aivenSecretFolders = append(aivenSecretFolders, AivenSecretFolder{path})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking path: %w", err)
	}
	return aivenSecretFolders, nil
}
