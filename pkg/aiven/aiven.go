package aiven

import (
	"context"
	"fmt"
	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	"github.com/nais/liberator/pkg/namegen"
	aivenclient "github.com/nais/nais-d/pkg/client"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	AivenApiVersion          = "aiven.nais.io/v1"
	AivenKind                = "aivenApplication"
	DefaultProtected         = true
	MaxServiceUserNameLength = 64
)

type AivenConfiguration struct {
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

func SetupAivenConfiguration(properties AivenProperties) *AivenConfiguration {
	return &AivenConfiguration{properties}
}

func (a *AivenConfiguration) GenerateApplication() error {
	ctx := context.Background()
	client := aivenclient.SetupClient()

	namespace := v1.Namespace{}
	err := client.Get(ctx, kubeclient.ObjectKey{
		Namespace: a.Namespace,
		Name:      a.Namespace,
	}, &namespace)
	if err != nil {
		return err
	}
	a.Namespace = namespace.Name

	timeStamp := time.Now().AddDate(0, 0, a.Expiry).Format(time.RFC3339)
	aivenApp := *a.CreateAivenApplication(timeStamp, a.SecretName)

	existingAivenApp := aiven_nais_io_v1.AivenApplication{}
	err = client.Get(ctx, kubeclient.ObjectKey{
		Namespace: a.Namespace,
		Name:      a.Username,
	}, &existingAivenApp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			fmt.Printf("Creating aivenApp %s\n", aivenApp.Name)
			err = client.Create(ctx, &aivenApp)
		}
	} else {
		fmt.Printf("Updating aivenApp %s\n", existingAivenApp.Name)
		aivenApp.ResourceVersion = existingAivenApp.ResourceVersion
		err = client.Update(ctx, &aivenApp)
	}

	if err != nil {
		return err
	}

	fmt.Printf("To get secret and config run cmd --> 'nais-d aiven get %s %s -c kcat'", aivenApp.Spec.SecretName, a.Namespace)
	return nil
}

func (c AivenConfiguration) CreateAivenApplication(timeStamp, secretName string) *aiven_nais_io_v1.AivenApplication {
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
	err := SetSecretName(app, secretName)
	if err != nil {
		return nil
	}
	return app
}

func SetSecretName(aivenApp *aiven_nais_io_v1.AivenApplication, secretName string) error {
	if secretName != "" {
		aivenApp.Spec.SecretName = secretName
	} else {
		newSecretName, err := setSecretName(aivenApp)
		if err != nil {
			return fmt.Errorf("could not create secretName: %s", err)
		}
		aivenApp.Spec.SecretName = newSecretName
	}
	return nil
}

func setSecretName(aivenApp *aiven_nais_io_v1.AivenApplication) (string, error) {
	return namegen.ShortName(SecretNamePrefix(aivenApp.Namespace, aivenApp.Name), MaxServiceUserNameLength)
}

func SecretNamePrefix(username, team string) string {
	return fmt.Sprintf("%s-%s", team, username)
}
