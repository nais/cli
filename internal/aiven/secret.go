package aiven

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nais/cli/internal/aiven/aiven_config"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/liberator/pkg/namegen"
	"github.com/nais/naistrix"
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

var ErrUnsuitableSecret = errors.New("unsuitable secret")

func ExtractAndGenerateConfig(ctx context.Context, service aiven_services.Service, secretName, namespaceName string, out *naistrix.OutputWriter) error {
	aivenClient := k8s.SetupControllerRuntimeClient()

	if err := validateNamespace(ctx, aivenClient, namespaceName); err != nil {
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
	if !hasAnnotation(existingSecret, AivenatorProtectedAnnotation) && !hasAnnotation(existingSecret, AivenatorProtectedExpireAtAnnotation) {
		return fmt.Errorf("secret is must have at least one of these annotations(%w): '%s', '%s'", ErrUnsuitableSecret, AivenatorProtectedAnnotation, AivenatorProtectedExpireAtAnnotation)
	}

	if err := secret.generateConfig(out); err != nil {
		return fmt.Errorf("generating config: %w", err)
	}

	if secret.Service.Is(&aiven_services.OpenSearch{}) {
		data := secret.Secret.Data
		out.Printf("OpenSearch dashboard: https://%s (username: %s, password: %s)",
			data[aiven_config.OpenSearchHostKey], data[aiven_config.OpenSearchUsernameKey], data[aiven_config.OpenSearchPasswordKey])
	}

	out.Printf("Configurations from secret '%s' found here:\n%s", existingSecret.Name, dest)
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
	if err := aiven_config.NewJavaConfig(s.Secret, s.DestinationPath); err != nil {
		return err
	}

	if err := aiven_config.WriteKCatConfigToFile(s.Secret, s.DestinationPath); err != nil {
		return err
	}

	if err := aiven_config.WriteKafkaEnvConfigToFile(s.Secret, s.DestinationPath); err != nil {
		return err
	}

	return nil
}

func (s *Secret) CreateOpenSearchConfigs() error {
	return aiven_config.WriteOpenSearchEnvConfigToFile(s.Secret, s.DestinationPath)
}

func (s *Secret) generateConfig(out *naistrix.OutputWriter) error {
	out.Printf("Generating %v config from secret %v", s.Service.Name(), s.Secret.Name)
	return s.Service.Generate(s)
}

func createDefaultDestination() (string, error) {
	newPath, err := os.MkdirTemp("", FolderPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	return newPath, nil
}

func CreateSecretName(username, namespace string) (string, error) {
	return namegen.ShortName(
		fmt.Sprintf("%s-%s", username, strings.ReplaceAll(namespace, ".", "-")),
		64,
	)
}
