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

func ExtractAndGenerateConfig(configTyp, dest, secretName, team string) {
	aivenClient := client.SetupClient()
	ctx := context.Background()

	dest, err := helpers.DefaultDestination(dest)
	if err != nil {
		log.Fatalf("an error %s", err)
	}

	namespace := v1.Namespace{}
	err = common.ValidateNamespace(ctx, aivenClient, team, &namespace)
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

	err = Config(secret, dest, configTyp)
	if err != nil {
		log.Fatalf("an error %s", err)
	}
}

func hasAnnotation(secret *v1.Secret, key string) bool {
	if value, ok := secret.GetAnnotations()[key]; ok && value == "true" {
		return ok
	}
	return false
}

func ConfigAll(secret *v1.Secret, dest string) error {
	kafkaEnv := config.NewEnvConfig(secret, dest)
	kafkaEnv.Init()
	kCatConfig := config.NewKCatConfig(secret, dest)
	kCatConfig.Init()
	err := kafkaEnv.Generate()
	if err != nil {
		return err
	}
	err = kCatConfig.Generate()
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

func Config(secret *v1.Secret, dest, typeConfig string) error {
	switch typeConfig {

	case config.ENV:
		kafkaEnv := config.NewEnvConfig(secret, dest)
		kafkaEnv.Init()
		err := kafkaEnv.Generate()
		if err != nil {
			return err
		}

		if err := kafkaEnv.Finit(); err != nil {
			return err
		}
		log.Default().Printf("%s and secrets (%s) generated: %s.", typeConfig, secret.Name, dest)
	case config.KCAT:
		kCatConfig := config.NewKCatConfig(secret, dest)
		kCatConfig.Init()
		err := kCatConfig.Generate()
		if err != nil {
			return err
		}

		if err := kCatConfig.Finit(); err != nil {
			return err
		}
		log.Default().Printf("%s and secrets (%s) generated: %s.", typeConfig, secret.Name, dest)
	case config.ALL:
		err := ConfigAll(secret, dest)
		if err != nil {
			return fmt.Errorf("generate all configs: %s", err)
		}
		log.Default().Printf("%s configs and secrets (%s) generated: %s.", typeConfig, secret.Name, dest)
	}
	return nil
}
