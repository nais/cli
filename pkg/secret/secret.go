package secret

import (
	"context"
	"fmt"
	"github.com/nais/nais-d/pkg/client"
	"github.com/nais/nais-d/pkg/config"
	"github.com/nais/nais-d/pkg/consts"
	v1 "k8s.io/api/core/v1"
	"os"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AivenatorProtectedAnnotation         = "aivenator.aiven.nais.io/protected"
	AivenatorProtectedExpireAtAnnotation = "aivenator.aiven.nais.io/with-time-limit"
)

func ExtractAndGenerateConfig(configTyp, dest, secretName, team string) {
	aivenClient := client.SetupClient()
	ctx := context.Background()

	namespace := v1.Namespace{}
	err := aivenClient.Get(ctx, ctrl.ObjectKey{
		Name: team,
	}, &namespace)
	if err != nil {
		fmt.Printf("an error %s", err)
		os.Exit(1)
	}

	secret := &v1.Secret{}
	err = aivenClient.Get(ctx, ctrl.ObjectKey{
		Namespace: namespace.Name,
		Name:      secretName,
	}, secret)
	if err != nil {
		fmt.Printf("an error %s", err)
		os.Exit(1)
	}

	// check is annotations match with protected and time-limited otherwise you could use any secret!
	if !hasAnnotation(secret, AivenatorProtectedAnnotation) || !hasAnnotation(secret, AivenatorProtectedExpireAtAnnotation) {
		fmt.Printf("secret is missing annotations: '%s' and '%s'", AivenatorProtectedAnnotation, AivenatorProtectedExpireAtAnnotation)
		os.Exit(1)
	}

	err = Config(secret, dest, configTyp)
	if err != nil {
		fmt.Printf("an error %s", err)
		os.Exit(1)
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
	kCatConfig.Generate()
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

	case consts.ENV:
		kafkaEnv := config.NewEnvConfig(secret, dest)
		kafkaEnv.Init()
		err := kafkaEnv.Generate()
		if err != nil {
			return err
		}

		if err := kafkaEnv.Finit(); err != nil {
			return err
		}
	case consts.KCAT:
		kCatConfig := config.NewKCatConfig(secret, dest)
		kCatConfig.Init()
		err := kCatConfig.Generate()
		if err != nil {
			return err
		}

		if err := kCatConfig.Finit(); err != nil {
			return err
		}

	case consts.ALL:
		err := ConfigAll(secret, dest)
		if err != nil {
			return fmt.Errorf("generate all configs: %s", err)
		}
	}
	return nil
}
