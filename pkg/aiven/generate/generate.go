package generate

import (
	"fmt"
	"github.com/nais/debuk/pkg/aiven/application"
	"github.com/nais/debuk/pkg/aiven/secret"
	"github.com/nais/debuk/pkg/command"
	"gopkg.in/yaml.v3"
	"time"
)

func AivenApplication(username, team, pool, dest string, expire int, secretName string) error {
	fmt.Printf("destination folder is set to --> %s\n", dest)

	timeStamp := time.Now().AddDate(0, 0, expire).Format(time.RFC3339)

	aiven := application.CreateAiven(username, team, pool, timeStamp)

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

func setExpiry(expire string) (string, error) {
	if expire == "" {
		expire = time.Now().AddDate(0, 0, 1).Format(time.RFC3339)
	} else {
		parsedTimeStamp, err := time.Parse(time.RFC3339, expire)
		if err != nil {
			return "", fmt.Errorf("could not parse expire: %s", err)
		}
		return parsedTimeStamp.String(), nil
	}
	return expire, nil
}
