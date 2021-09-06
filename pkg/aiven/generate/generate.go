package generate

import (
	"fmt"
	"github.com/nais/debuk/pkg/aiven/application"
	"github.com/nais/debuk/pkg/aiven/secret"
	"github.com/nais/debuk/pkg/command"
	"gopkg.in/yaml.v3"
)

func AivenApplication(username, team, pool, dest string, expire int, secretName string) error {
	fmt.Printf("destination folder is set to --> %s\n", dest)

	aiven := application.CreateAiven(username, team, pool, expire)

	if err := aiven.SetSecretName(secretName); err != nil {
		return err
	}

	aivenYamlPath := aiven.PathToFile(username, team, dest)
	if err := aiven.MarshalAndWriteToFile(aivenYamlPath); err != nil {
		return err
	}

	if _, err := command.Apply(aivenYamlPath); err != nil {
		return err
	}

	stdoutSecret, err := command.GetSecret(aiven.Spec.SecretName)
	if err != nil {
		return err
	}

	receivedSecret := secret.Secret{}
	if err := yaml.Unmarshal(stdoutSecret, &receivedSecret); err != nil {
		return err
	}

	if err = receivedSecret.GenerateConfiguration(dest, username); err != nil {
		return err
	}

	fmt.Printf("Debuked! Files found here --> %s/*", dest)
	return nil
}
