package aiven

import (
	"fmt"
	"os"

	"github.com/nais/cli/internal/aiven/aiven_services"
	v1 "k8s.io/api/core/v1"
)

const (
	FolderPrefix = "aiven-secret-"
)

type Secret struct {
	Secret          *v1.Secret
	DestinationPath string
	Service         aiven_services.Service
}

func createDefaultDestination() (string, error) {
	newPath, err := os.MkdirTemp("", FolderPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	return newPath, nil
}
