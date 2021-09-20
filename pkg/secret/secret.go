package secret

import (
	"context"
	"fmt"
	"github.com/nais/nais-cli/cmd/helpers"
	"github.com/nais/nais-cli/pkg/client"
	"github.com/nais/nais-cli/pkg/common"
	"github.com/nais/nais-cli/pkg/config"
	v1 "k8s.io/api/core/v1"
	"log"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AivenatorProtectedAnnotation         = "aivenator.aiven.nais.io/protected"
	AivenatorProtectedExpireAtAnnotation = "aivenator.aiven.nais.io/with-time-limit"
)

func ExtractAndGenerateConfig(configTyp, dest, secretName, namespaceName string) {
	aivenClient := client.SetupClient()
	ctx := context.Background()

	dest, err := helpers.DefaultDestination(dest)
	if err != nil {
		log.Fatalf("an error %s", err)
	}

	namespace := v1.Namespace{}
	err = common.ValidateNamespace(ctx, aivenClient, namespaceName, &namespace)
	if err != nil {
		log.Fatalf("an error %s", err)
	}

	secret := &v1.Secret{}
	err = aivenClient.Get(ctx, ctrl.ObjectKey{
		Namespace: namespace.Name,
		Name:      secretName,
	}, secret)
	if err != nil {
		log.Fatalf("an error %s", err)
	}

	// check is annotations match with protected and time-limited otherwise you could use any secret!
	if !hasAnnotation(secret, AivenatorProtectedAnnotation) || !hasAnnotation(secret, AivenatorProtectedExpireAtAnnotation) {
		log.Fatalf("secret is missing annotations: '%s' or '%s'", AivenatorProtectedAnnotation, AivenatorProtectedExpireAtAnnotation)
	}

	_, err = Config(secret, dest, configTyp)
	if err != nil {
		log.Fatalf("an error %s", err)
	}
	log.Default().Printf("'%s' config from secret '%s' found her: '%s'.", configTyp, secret.Name, dest)
}

func hasAnnotation(secret *v1.Secret, key string) bool {
	if value, ok := secret.GetAnnotations()[key]; ok && value == "true" {
		return ok
	}
	return false
}

func ConfigAll(secret *v1.Secret, dest string) error {
	kafkaEnv := config.NewEnvConfig(secret, config.KafkaConfigEnvToFileMap, dest)
	kafkaEnv.Init()
	kCatConfig := config.NewKCatConfig(secret, config.KCatEnvToFileMap, dest)
	kCatConfig.Init()
	_, err := kafkaEnv.Generate()
	if err != nil {
		return err
	}
	_, err = kCatConfig.Generate()
	if err != nil {
		return err
	}

	if err := kCatConfig.Finit(); err != nil {
		return err
	}

	if err := kafkaEnv.Finit(); err != nil {
		return err
	}
	return nil
}

func Config(secret *v1.Secret, dest, typeConfig string) (string, error) {
	log.Default().Printf("generating '%s' from secret '%s'", typeConfig, secret.Name)
	switch typeConfig {
	case config.ENV:
		kafkaEnv := config.NewEnvConfig(secret, config.KafkaConfigEnvToFileMap, dest)
		kafkaEnv.Init()
		envs, err := kafkaEnv.Generate()
		if err != nil {
			return "", err
		}

		if err := kafkaEnv.Finit(); err != nil {
			return "", err
		}
		return envs, nil
	case config.KCAT:
		kCatConfig := config.NewKCatConfig(secret, config.KCatEnvToFileMap, dest)
		kCatConfig.Init()
		kCat, err := kCatConfig.Generate()
		if err != nil {
			return "", err
		}

		if err := kCatConfig.Finit(); err != nil {
			return "", err
		}
		return kCat, nil
	case config.ALL:
		err := ConfigAll(secret, dest)
		if err != nil {
			return "", fmt.Errorf("generate all configs: %s", err)
		}
	}
	return "", nil
}
