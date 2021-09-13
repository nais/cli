package aiven

import (
	"fmt"
	"github.com/nais/nais-d/client"
	"github.com/nais/nais-d/pkg/consts"
	"github.com/nais/nais-d/pkg/secrets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

func TypeConfig(configTyp, dest, secretName, team string) {
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

	if configTyp == consts.ENV || configTyp == consts.KCAT {
		err = secrets.Config(secret, dest, configTyp)
		if err != nil {
			fmt.Printf("an error %s", err)
			os.Exit(1)
		}
	} else {
		err = secrets.GenerateAll(secret, dest)
		if err != nil {
			fmt.Printf("an error %s", err)
			os.Exit(1)
		}
	}
}
