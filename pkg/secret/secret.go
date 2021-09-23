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

func GetExistingSecret(ctx context.Context, client ctrl.Client, namespace, secretName string) *v1.Secret {
	secret := &v1.Secret{}
	err := client.Get(ctx, ctrl.ObjectKey{
		Namespace: namespace,
		Name:      secretName,
	}, secret)
	if err != nil {
		log.Fatalf("an error %s", err)
	}
	return secret
}

func ExtractAndGenerateConfig(configTyp, dest, secretName, namespaceName string) {
	aivenClient := client.SetupClient()
	ctx := context.Background()

	dest, err := cmd.DefaultDestination(dest)
	if err != nil {
		log.Fatalf("an error %s", err)
	}

	namespace := v1.Namespace{}
	err = common.ValidateNamespace(ctx, aivenClient, namespaceName, &namespace)
	if err != nil {
		log.Fatalf("an error %s", err)
	}

	existingSecret := GetExistingSecret(ctx, aivenClient, namespace.Name, secretName)
	secret := SetupSecretConfiguration(existingSecret, configTyp, dest)

	// check is annotations match with protected and time-limited otherwise you could use any existingSecret!
	if !hasAnnotation(existingSecret, AivenatorProtectedAnnotation) || !hasAnnotation(existingSecret, AivenatorProtectedExpireAtAnnotation) {
		log.Fatalf("existingSecret is missing annotations: '%s' or '%s'", AivenatorProtectedAnnotation, AivenatorProtectedExpireAtAnnotation)
	}

	_, err = secret.Config()
	if err != nil {
		log.Fatalf("an error %s", err)
	}
	log.Default().Printf("configurations from secret '%s' found her: '%s'.", existingSecret.Name, dest)
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
