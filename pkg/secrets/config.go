package secrets

import (
	"context"
	"fmt"
	"github.com/nais/nais-d/pkg/client"
	"github.com/nais/nais-d/pkg/consts"
	v1 "k8s.io/api/core/v1"
	"os"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
)

func ExtractAndGenerateConfig(configTyp, dest, secretName, team string) {
	stdClient := client.SetupClient()
	ctx := context.Background()
	key := client2.ObjectKey{
		Namespace: team,
		Name:      team,
	}

	namespace := v1.Namespace{}
	err := stdClient.Get(ctx, key, &namespace)
	if err != nil {
		fmt.Printf("an error %s", err)
		os.Exit(1)
	}

	key2 := client2.ObjectKey{
		Namespace: namespace.Name,
		Name:      secretName,
	}

	secret := v1.Secret{}
	err = stdClient.Get(ctx, key2, &secret)
	if err != nil {
		fmt.Printf("an error %s", err)
		os.Exit(1)
	}

	// check is annotations match with protected and time-limited otherwise you could use any secret!
	if !hasAnnotation(&secret, AivenatorProtectedAnnotation) || !hasAnnotation(&secret, AivenatorProtectedExpireAtAnnotation) {
		fmt.Printf("secret is missing annotations: '%s' and '%s'", AivenatorProtectedAnnotation, AivenatorProtectedExpireAtAnnotation)
		os.Exit(1)
	}

	if configTyp == consts.ENV || configTyp == consts.KCAT {
		err = Config(&secret, dest, configTyp)
		if err != nil {
			fmt.Printf("an error %s", err)
			os.Exit(1)
		}
	} else {
		err = GenerateAll(&secret, dest)
		if err != nil {
			fmt.Printf("an error %s", err)
			os.Exit(1)
		}
	}
}

func hasAnnotation(secret *v1.Secret, key string) bool {
	if value, ok := secret.GetAnnotations()[key]; ok && value == "true" {
		return ok
	}
	return false
}
