package secrets

import (
	"fmt"
	"github.com/nais/nais-d/pkg/client"
	"github.com/nais/nais-d/pkg/consts"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

func ExtractAndGenerateConfig(configTyp, dest, secretName, team string) {
	stdClient := client.StandardClient()

	namespace, err := stdClient.CoreV1().Namespaces().Get(team, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("an error %s", err)
		os.Exit(1)
	}

	secret, err := stdClient.CoreV1().Secrets(namespace.Name).Get(secretName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("an error %s", err)
		os.Exit(1)
	}

	// check is annotations match with protected and time-limited otherwise you could use any secret!
	if !hasAnnotation(secret, AivenatorProtectedAnnotation) || !hasAnnotation(secret, AivenatorProtectedExpireAtAnnotation) {
		fmt.Printf("secret is missing annotations: '%s' and '%s'", AivenatorProtectedAnnotation, AivenatorProtectedExpireAtAnnotation)
		os.Exit(1)
	}

	if configTyp == consts.ENV || configTyp == consts.KCAT {
		err = Config(secret, dest, configTyp)
		if err != nil {
			fmt.Printf("an error %s", err)
			os.Exit(1)
		}
	} else {
		err = GenerateAll(secret, dest)
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
