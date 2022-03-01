package secret

import (
	"context"
	"fmt"
	"github.com/nais/cli/cmd"
	"github.com/nais/cli/pkg/aiven"
	"github.com/nais/cli/pkg/client"
	"github.com/nais/cli/pkg/common"
	"github.com/nais/cli/pkg/config"
	v1 "k8s.io/api/core/v1"
	"log"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AivenatorProtectedAnnotation         = "aivenator.aiven.nais.io/protected"
	AivenatorProtectedExpireAtAnnotation = "aivenator.aiven.nais.io/with-time-limit"
)

type Secret struct {
	Secret          *v1.Secret
	DestinationPath string
	Service         aiven.Service
}

func SetupSecretConfiguration(secret *v1.Secret, dest string, service aiven.Service) Secret {
	return Secret{
		Secret:          secret,
		DestinationPath: dest,
		Service:         service,
	}
}

func GetExistingSecret(ctx context.Context, client ctrl.Client, namespace, secretName string) (*v1.Secret, error) {
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

func ExtractAndGenerateConfig(service aiven.Service, secretName, namespaceName string) error {
	aivenClient := client.SetupClient()
	ctx := context.Background()

	namespace := v1.Namespace{}
	err := common.ValidateNamespace(ctx, aivenClient, namespaceName, &namespace)
	if err != nil {
		return fmt.Errorf("validate namespace: %w", err)
	}

	dest, err := cmd.DefaultDestination()
	if err != nil {
		return fmt.Errorf("setting default folder: %w", err)
	}

	existingSecret, err := GetExistingSecret(ctx, aivenClient, namespace.Name, secretName)
	if err != nil {
		return err
	}

	secret := SetupSecretConfiguration(existingSecret, dest, service)

	// check if annotations match with protected or time-limited otherwise you could use any existingSecret!
	if !(hasAnnotation(existingSecret, AivenatorProtectedAnnotation) || hasAnnotation(existingSecret, AivenatorProtectedExpireAtAnnotation)) {
		return fmt.Errorf("secret is must have at least one of these annotations: '%s', '%s'", AivenatorProtectedAnnotation, AivenatorProtectedExpireAtAnnotation)
	}

	err = secret.Config()
	if err != nil {
		return fmt.Errorf("generating config: %w", err)
	}
	log.Default().Printf("configurations from secret '%s' found here: '%s'.", existingSecret.Name, dest)
	return nil
}

func hasAnnotation(secret *v1.Secret, key string) bool {
	if value, ok := secret.GetAnnotations()[key]; ok && value == "true" {
		return ok
	}
	return false
}

func (s *Secret) CreateKafkaConfigs() error {
	err := config.NewJavaConfig(s.Secret, s.DestinationPath)
	if err != nil {
		return err
	}
	err = config.WriteKCatConfigToFile(s.Secret, s.DestinationPath)
	if err != nil {
		return err
	}
	err = config.WriteKafkaEnvConfigToFile(s.Secret, s.DestinationPath)
	if err != nil {
		return err
	}
	return nil
}

func (s *Secret) Config() error {
	log.Default().Printf("generating %v config from secret %v", s.Service, s.Secret.Name)
	switch s.Service {
	case aiven.Kafka:
		err := s.CreateKafkaConfigs()
		if err != nil {
			return err
		}
	case aiven.OpenSearch:
		err := config.WriteOpenSearchEnvConfigToFile(s.Secret, s.DestinationPath)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unkown service: %v", s.Service)
	}
	return nil
}
