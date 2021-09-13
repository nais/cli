package config

import (
	"fmt"
	aivenclient "github.com/nais/debuk/client"
	"github.com/nais/debuk/cmd/helpers"
	"github.com/nais/debuk/pkg/application"
	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (
	AivenApiVersion  = "aiven.nais.io/v1"
	AivenKind        = "aivenApplication"
	DefaultProtected = true
)

type AivenConfiguration struct {
	client *kubernetes.Clientset
	AivenProperties
}

type AivenProperties struct {
	Username   string
	Namespace  string
	Pool       string
	Dest       string
	SecretName string
	Expiry     int
}

func SetupAivenConfiguration(client *kubernetes.Clientset, properties AivenProperties) *AivenConfiguration {
	return &AivenConfiguration{
		client,
		properties,
	}
}

func (a *AivenConfiguration) GenerateApplication() error {
	dest, err := helpers.DefaultDestination(a.Dest)
	if err != nil {
		return fmt.Errorf("setting destination: %s", err)
	}
	fmt.Printf("destination folder is set to --> %s\n", dest)

	timeStamp := time.Now().AddDate(0, 0, a.Expiry).Format(time.RFC3339)
	aiven := application.CreateAiven(a.Username, a.Namespace, a.Pool, timeStamp)
	if err := aiven.SetSecretName(a.SecretName); err != nil {
		return err
	}

	aivenYamlPath := aiven.PathToFile(a.Username, a.Namespace, dest)
	if err := aiven.MarshalAndWriteToFile(aivenYamlPath); err != nil {
		return err
	}

	client, err := aivenclient.NewForConfig()
	if err != nil {
		panic(err)
	}

	update := true
	getAivenApp, err := client.Aiven(a.Namespace).Get(a.Username, metav1.GetOptions{})
	createOrUpdateAivenApp := a.CreateAiven(timeStamp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			createdApp, err := client.Aiven(a.Namespace).Create(createOrUpdateAivenApp)
			if err != nil {
				return err
			}
			update = false
			fmt.Printf("app: %s created\n", createdApp.Name)
		} else {
			return err
		}
	}

	if update {
		createOrUpdateAivenApp.SetResourceVersion(getAivenApp.ResourceVersion)
		updatedApp, err := client.Aiven(a.Namespace).Update(createOrUpdateAivenApp)
		if err != nil {
			return err
		}
		fmt.Printf("configured app: %s\n", updatedApp.Name)
	}

	fmt.Printf("Debuked! AivenApplication: %s found here --> %s/*\n", aiven.Metadata.Name, dest)
	fmt.Printf("Get secrets and configs run cmd --> debuk get -c kcat -s %s -d %s", aiven.Spec.SecretName, a.Dest)
	return nil
}

func (c AivenConfiguration) CreateAiven(timeStamp string) *aiven_nais_io_v1.AivenApplication {
	app := &aiven_nais_io_v1.AivenApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       AivenKind,
			APIVersion: AivenApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Username,
			Namespace: c.Namespace,
		},
		Spec: aiven_nais_io_v1.AivenApplicationSpec{
			SecretName: "",
			Protected:  DefaultProtected,
			ExpiresAt:  timeStamp,
			Kafka:      &aiven_nais_io_v1.KafkaSpec{Pool: c.Pool},
		},
	}
	return app
}
