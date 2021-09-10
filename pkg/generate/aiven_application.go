package generate

import (
	"fmt"
	"github.com/nais/debuk/cmd/helpers"
	"github.com/nais/debuk/pkg/application"
	"github.com/nais/debuk/pkg/kubectl"
	"time"
)

func AivenApplication(username, team, pool, originalDest string, expire int, secretName string) error {
	dest, err := helpers.DefaultDestination(originalDest)
	if err != nil {
		return fmt.Errorf("setting destination: %s", err)
	}
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

	if _, err := kubectl.Apply(aivenYamlPath); err != nil {
		return err
	}

	fmt.Printf("Debuked! AivenApplication: %s found here --> %s/*\n", aiven.Metadata.Name, dest)
	fmt.Printf("To get secrets and generate config run cmd --> debuk get -c kcat -s %s -d %s", aiven.Spec.SecretName, originalDest)
	return nil
}
