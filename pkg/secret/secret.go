package secret

import (
	"context"
	"fmt"
	"github.com/nais/cli/cmd"
	"github.com/nais/cli/pkg/aiven"
	"github.com/nais/cli/pkg/client"
	"github.com/nais/cli/pkg/common"
	"github.com/nais/cli/pkg/config"
	"github.com/nais/cli/pkg/consts"
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
	ConfigType      string
	DestinationPath string
	Service         aiven.Service
}

func SetupSecretConfiguration(secret *v1.Secret, configType, dest string, service aiven.Service) Secret {
	return Secret{
		Secret:          secret,
		ConfigType:      configType,
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

func ExtractAndGenerateConfig(service aiven.Service, configType, secretName, namespaceName string) error {
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

	secret := SetupSecretConfiguration(existingSecret, configType, dest, service)

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

func (s *Secret) CreateAllConfigs() error {
	if err := s.CreateJavaConfig(); err != nil {
		return err
	}
	if err := s.CreateKCatConfig(); err != nil {
		return err
	}
	if err := s.CreateEnvConfig(); err != nil {
		return err
	}
	return nil
}

func createKafkaSecrets(s *Secret) error {

	switch s.ConfigType {
	case consts.EnvironmentConfigurationType:
		return s.CreateEnvConfig()
	case consts.KCatConfigurationType:
		return s.CreateKCatConfig()
	case consts.JavaConfigurationType:
		return s.CreateJavaConfig()
	case consts.AllConfigurationType:

	}
	return nil
}

func (s *Secret) Config() error {
	log.Default().Printf("generating '%s' from secret '%s'", s.ConfigType, s.Secret.Name)
	switch s.Service {
	case aiven.Kafka:
		err := s.CreateAllConfigs()
		if err != nil {
			return fmt.Errorf("generate %s config-type", s.ConfigType)
		}
	default:
		return fmt.Errorf("unkown service: %v", s.Service)
	}
	return nil
}

func (s *Secret) CreateJavaConfig() error {
	javaConfig := config.NewJavaConfig(s.Secret, s.DestinationPath)
	_, err := javaConfig.Generate()
	if err != nil {
		return fmt.Errorf("generate %s config-type", s.ConfigType)
	}

	if err := javaConfig.WriteConfigToFile(); err != nil {
		return err
	}
	return nil
}

func (s *Secret) CreateKCatConfig() error {
	kCatConfig := config.NewKCatConfig(s.Secret, s.DestinationPath)
	_, err := kCatConfig.Generate()
	if err != nil {
		return fmt.Errorf("generate %s config-type", s.ConfigType)
	}

	if err := kCatConfig.WriteConfigToFile(); err != nil {
		return err
	}
	return nil
}

func (s *Secret) CreateEnvConfig() error {
	kafkaEnv := config.NewEnvConfig(s.Secret, s.DestinationPath)
	_, err := kafkaEnv.Generate()
	if err != nil {
		return fmt.Errorf("generate %s config-type", s.ConfigType)
	}

	if err := kafkaEnv.WriteConfigToFile(); err != nil {
		return err
	}
	return nil
}
