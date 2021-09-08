package application

import (
	"fmt"
	"github.com/nais/liberator/pkg/namegen"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

const (
	AivenApiVersion = "aiven.nais.io/v1"
	AivenKind       = "AivenApplication"

	DefaultProtected = true
	MaxServiceUserNameLength = 64
	FilePermission = 0775
)

type Aiven struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string
	Metadata   Metadata
	Spec       AivenSpec
}

type Metadata struct {
	Name      string
	Namespace string
}

type AivenSpec struct {
	SecretName string `yaml:"secretName"`
	Protected  bool
	Kafka      KafkaSpec
	ExpiresAt  string `yaml:"expiresAt"`
}

type KafkaSpec struct {
	Pool string
}

func CreateAiven(username, team, pool string, expiryDate string) Aiven {
	app := Aiven{
		Kind:       AivenKind,
		ApiVersion: AivenApiVersion,
		Metadata: Metadata{
			Name:      username,
			Namespace: team,
		},
		Spec: AivenSpec{
			Kafka: KafkaSpec{
				Pool: pool,
			},
			SecretName: "",
			Protected:  DefaultProtected,
			ExpiresAt:  expiryDate,
		},
	}
	return app
}

func (a *Aiven) SetSecretName(secretName string) error {
	if secretName != "" {
		a.Spec.SecretName = secretName
	} else {
		newSecretName, err := a.secretName()
		if err != nil {
			return fmt.Errorf("could not create secretName: %s", err)
		}
		a.Spec.SecretName = newSecretName
	}
	return nil
}

func (a *Aiven) secretName() (string, error) {
	return namegen.ShortName(SecretNamePrefix(a.Metadata.Namespace, a.Metadata.Name), MaxServiceUserNameLength)
}

func SecretNamePrefix(username, team string) string {
	return fmt.Sprintf("%s-%s", team, username)
}

func (a *Aiven) MarshalAndWriteToFile(dest string) error {
	yamlData, err := yaml.Marshal(&a)

	if err != nil {
		return fmt.Errorf("error while Marshaling. %v", err)
	}

	err = ioutil.WriteFile(dest, yamlData, FilePermission)
	if err != nil {
		return fmt.Errorf("unable to write data to file: %s", err)
	}
	return nil
}

func (a *Aiven) PathToFile(username, team, dest string) string {
	return fmt.Sprintf("%s/%s.yaml", dest, SecretNamePrefix(username, team))
}
