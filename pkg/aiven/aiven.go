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

type Aiven struct {
	Ctx    context.Context
	Client kubeclient.Client
	Props  AivenProperties
}

type AivenProperties struct {
	Username   string
	Namespace  string
	Pool       string
	Dest       string
	SecretName string
	Expiry     int
}

func SetupAiven(client kubeclient.Client, username, team, pool, secretName string, expiry int) *Aiven {
	return &Aiven{
		context.Background(),
		client,
		AivenProperties{
			Username:   username,
			Namespace:  team,
			Pool:       pool,
			SecretName: secretName,
			Expiry:     expiry,
		},
	}
}

func (a *Aiven) GenerateApplication() error {
	client := aivenclient.SetupClient()

	namespace := v1.Namespace{}
	err := client.Get(a.Ctx, kubeclient.ObjectKey{
		Name:      a.Props.Namespace,
	}, &namespace)
	if err != nil {
		return err
	}
	a.Props.Namespace = namespace.Name

	timeStamp := time.Now().AddDate(0, 0, a.Props.Expiry).Format(time.RFC3339)
	aivenApp := *a.CreateAivenApplication(timeStamp, a.Props.SecretName)

	existingAivenApp := aiven_nais_io_v1.AivenApplication{}
	err = client.Get(a.Ctx, kubeclient.ObjectKey{
		Namespace: a.Props.Namespace,
		Name:      a.Props.Username,
	}, &existingAivenApp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			fmt.Printf("Creating aivenApp %s\n", aivenApp.Name)
			err = client.Create(a.Ctx, &aivenApp)
		}
	} else {
		fmt.Printf("Updating aivenApp %s\n", existingAivenApp.Name)
		aivenApp.ResourceVersion = existingAivenApp.ResourceVersion
		err = client.Update(a.Ctx, &aivenApp)
	}

	if err != nil {
		return err
	}

	fmt.Printf("To get secret and config run cmd --> 'nais-d aiven get %s %s -c kcat'", aivenApp.Spec.SecretName, a.Props.Namespace)
	return nil
}

func (a Aiven) CreateAivenApplication(timeStamp, secretName string) *aiven_nais_io_v1.AivenApplication {
	app := &aiven_nais_io_v1.AivenApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       AivenKind,
			APIVersion: AivenApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Props.Username,
			Namespace: a.Props.Namespace,
		},
		Spec: aiven_nais_io_v1.AivenApplicationSpec{
			SecretName: "",
			Protected:  DefaultProtected,
			ExpiresAt:  timeStamp,
			Kafka:      &aiven_nais_io_v1.KafkaSpec{Pool: a.Props.Pool},
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
