package config

import (
	"fmt"
	"github.com/nais/debuk/pkg/consts"
	"github.com/nais/debuk/pkg/kubectl"
	"github.com/nais/debuk/pkg/secret"
	"gopkg.in/yaml.v3"
	"os"
)

func TypeConfig(configTyp, dest, secretName string) {
	stdoutSecretData, err := kubectl.GetSecret(secretName)
	if err != nil {
		fmt.Printf("an error %s", err)
		os.Exit(1)
	}

	receivedSecret := secret.Secret{}
	if err := yaml.Unmarshal(stdoutSecretData, &receivedSecret); err != nil {
		fmt.Printf("an error %s", err)
		os.Exit(1)
	}

	// check is annotations match with protected and time-limited otherwise you could use any secret!
	if configTyp == consts.ENV || configTyp == consts.KCAT {
		err = receivedSecret.Config(dest, configTyp)
		if err != nil {
			fmt.Printf("an error %s", err)
			os.Exit(1)
		}
	} else {
		err = receivedSecret.ConfigAll(dest)
		if err != nil {
			fmt.Printf("an error %s", err)
			os.Exit(1)
		}
	}
}
