package aiven

import (
	"context"
	"fmt"
	"github.com/nais/cli/pkg/k8s"
	"log"
	"os"

	"github.com/nais/cli/pkg/aiven/aiven_config"
	"github.com/nais/cli/pkg/aiven/aiven_services"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	FolderPrefix                         = "aiven-secret-"
	AivenatorProtectedAnnotation         = "aivenator.aiven.nais.io/protected"
	AivenatorProtectedExpireAtAnnotation = "aivenator.aiven.nais.io/with-time-limit"
)

type Secret struct {
	Secret          *v1.Secret
	DestinationPath string
	Service         aiven_services.Service
}

func ExtractAndGenerateConfig(service aiven_services.Service, secretName, namespaceName string) error {
	aivenClient := k8s.SetupClient()
	ctx := context.Background()

	err := validateNamespace(ctx, aivenClient, namespaceName)
	if err != nil {
		return fmt.Errorf("validate namespace: %w", err)
	}

	dest, err := createDefaultDestination()
	if err != nil {
		return fmt.Errorf("setting default folder: %w", err)
	}

	existingSecret, err := getExistingSecret(ctx, aivenClient, namespaceName, secretName)
	if err != nil {
		return err
	}

	secret := setupSecretConfiguration(existingSecret, dest, service)

	// check if annotations match with protected or time-limited otherwise you could use any existingSecret!
	if !(hasAnnotation(existingSecret, AivenatorProtectedAnnotation) || hasAnnotation(existingSecret, AivenatorProtectedExpireAtAnnotation)) {
		return fmt.Errorf("secret is must have at least one of these annotations: '%s', '%s'", AivenatorProtectedAnnotation, AivenatorProtectedExpireAtAnnotation)
	}

	err = secret.generateConfig()
	if err != nil {
		return fmt.Errorf("generating config: %w", err)
	}
	log.Default().Printf("configurations from secret '%s' found here: '%s'.", existingSecret.Name, dest)
	return nil
}

func setupSecretConfiguration(secret *v1.Secret, dest string, service aiven_services.Service) Secret {
	return Secret{
		Secret:          secret,
		DestinationPath: dest,
		Service:         service,
	}
}

func getExistingSecret(ctx context.Context, client ctrl.Client, namespace, secretName string) (*v1.Secret, error) {
	secret := &v1.Secret{}
	err := client.Get(ctx, ctrl.ObjectKey{
		Namespace: namespace,
		Name:      secretName,
	}, secret)
	if err != nil {
		return nil, fmt.Errorf("existing secret %w", err)
	}
	return secret, nil
}

func hasAnnotation(secret *v1.Secret, key string) bool {
	if value, ok := secret.GetAnnotations()[key]; ok && value == "true" {
		return ok
	}
	return false
}

func (s *Secret) CreateKafkaConfigs() error {
	err := aiven_config.NewJavaConfig(s.Secret, s.DestinationPath)
	if err != nil {
		return err
	}
	err = aiven_config.WriteKCatConfigToFile(s.Secret, s.DestinationPath)
	if err != nil {
		return err
	}
	err = aiven_config.WriteKafkaEnvConfigToFile(s.Secret, s.DestinationPath)
	if err != nil {
		return err
	}
	return nil
}

func (s *Secret) CreateOpenSearchConfigs() error {
	return aiven_config.WriteOpenSearchEnvConfigToFile(s.Secret, s.DestinationPath)
}

func (s *Secret) generateConfig() error {
	log.Default().Printf("generating %v config from secret %v", s.Service.Name(), s.Secret.Name)
	return s.Service.Generate(s)
}

func createDefaultDestination() (string, error) {
	newPath, err := os.MkdirTemp("", FolderPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	return newPath, nil
}
