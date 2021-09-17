package aiven

import (
	"context"
	"fmt"
	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	"github.com/nais/nais-cli/pkg/common"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	AivenApiVersion  = "aiven.nais.io/v1"
	AivenKind        = "aivenApplication"
	DefaultProtected = true
)

type Aiven struct {
	Ctx    context.Context
	Client ctrl.Client
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

func SetupAiven(client ctrl.Client, username, team, pool, secretName string, expiry int) *Aiven {
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

func (a *Aiven) GenerateApplication() (*aiven_nais_io_v1.AivenApplication, error) {
	namespace := v1.Namespace{}
	err := common.ValidateNamespace(a.Ctx, a.Client, a.Props.Namespace, &namespace)
	if err != nil {
		return nil, err
	}
	a.Props.Namespace = namespace.Name

	secretName, err := common.SetSecretName(a.Props.SecretName, a.Props.Username, a.Props.Namespace)
	aivenApp := *a.AivenApplication(secretName)

	err = a.CreateOrUpdate(&aivenApp)

	if err != nil {
		return nil, fmt.Errorf("create/update: %s", err)
	}
	return &aivenApp, nil
}

func (a Aiven) AivenApplication(secretName string) *aiven_nais_io_v1.AivenApplication {
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
			SecretName: secretName,
			Protected:  DefaultProtected,
			ExpiresAt:  time.Now().AddDate(0, 0, a.Props.Expiry).Format(time.RFC3339),
			Kafka:      &aiven_nais_io_v1.KafkaSpec{Pool: a.Props.Pool},
		},
	}
	return app
}

func (a Aiven) getExisting(existingAivenApp *aiven_nais_io_v1.AivenApplication) error {
	return a.Client.Get(a.Ctx, ctrl.ObjectKey{
		Namespace: a.Props.Namespace,
		Name:      a.Props.Username,
	}, existingAivenApp)
}

func (a Aiven) CreateOrUpdate(aivenApp *aiven_nais_io_v1.AivenApplication) error {
	existingAivenApp := aiven_nais_io_v1.AivenApplication{}
	err := a.getExisting(&existingAivenApp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			err = a.Client.Create(a.Ctx, aivenApp)
			if err != nil {
				return err
			}
			log.Default().Printf("aiven application: %s created.", aivenApp.Name)
		}
	} else {
		aivenApp.SetResourceVersion(existingAivenApp.GetResourceVersion())
		err = a.Client.Update(a.Ctx, aivenApp)
		if err != nil {
			return err
		}
		log.Default().Printf("aiven application: %s updated.", aivenApp.Name)
	}
	return nil
}
