package secret

import (
	"context"
	"fmt"
	"github.com/nais/nais-cli/cmd"
	"github.com/nais/nais-cli/pkg/client"
	"github.com/nais/nais-cli/pkg/common"
	"github.com/nais/nais-cli/pkg/config"
	"github.com/nais/nais-cli/pkg/consts"
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
}

func SetupSecretConfiguration(secret *v1.Secret, configType, dest string) Secret {
	return Secret{
		Secret:          secret,
		ConfigType:      configType,
		DestinationPath: dest,
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

func ExtractAndGenerateConfig(configTyp, secretName, namespaceName string) error {
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

	secret := SetupSecretConfiguration(existingSecret, configTyp, dest)

	// check is annotations match with protected and time-limited otherwise you could use any existingSecret!
	if !hasAnnotation(existingSecret, AivenatorProtectedAnnotation) || !hasAnnotation(existingSecret, AivenatorProtectedExpireAtAnnotation) {
		return fmt.Errorf("secret is missing annotations: '%s', '%s'", AivenatorProtectedAnnotation, AivenatorProtectedExpireAtAnnotation)
	}

	_, err = secret.Config()
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

func (s *Secret) ConfigAll() error {
	kafkaEnv := config.NewEnvConfig(s.Secret, s.DestinationPath)
	kCatConfig := config.NewKCatConfig(s.Secret, s.DestinationPath)
	_, err := kafkaEnv.Generate()
	if err != nil {
		return err
	}
	_, err = kCatConfig.Generate()
	if err != nil {
		return err
	}

	if err := kCatConfig.WriteConfigToFile(); err != nil {
		return err
	}

	if err := kafkaEnv.WriteConfigToFile(); err != nil {
		return err
	}
	return nil
}

func (s *Secret) Config() (string, error) {
	log.Default().Printf("generating '%s' from secret '%s'", s.ConfigType, s.Secret.Name)
	switch s.ConfigType {
	case consts.EnvironmentConfigurationType:
		kafkaEnv := config.NewEnvConfig(s.Secret, s.DestinationPath)
		envs, err := kafkaEnv.Generate()
		if err != nil {
			return "", fmt.Errorf("generate %s config-type", s.ConfigType)
		}

		if err := kafkaEnv.WriteConfigToFile(); err != nil {
			return "", err
		}
		return envs, nil
	case consts.KCatConfigurationType:
		kCatConfig := config.NewKCatConfig(s.Secret, s.DestinationPath)
		kCat, err := kCatConfig.Generate()
		if err != nil {
			return "", fmt.Errorf("generate %s config-type", s.ConfigType)
		}

		if err := kCatConfig.WriteConfigToFile(); err != nil {
			return "", err
		}
		return kCat, nil
	case consts.AllConfigurationType:
		err := s.ConfigAll()
		if err != nil {
			return "", fmt.Errorf("generate %s config-type", s.ConfigType)
		}
	}
	return "", nil
}
