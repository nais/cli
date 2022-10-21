package aiven

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	services2 "github.com/nais/cli/pkg/aiven/services"
	"github.com/nais/cli/pkg/common"
)

const (
	DefaultProtected = true
)

type Aiven struct {
	Ctx        context.Context
	Client     ctrl.Client
	Properties Properties
}

type Properties struct {
	Username   string
	Namespace  string
	Dest       string
	SecretName string
	Expiry     int
	Service    services2.Service
}

func Setup(innClient ctrl.Client, service services2.Service, username, namespace, secretName, instance string, pool services2.KafkaPool, access services2.OpenSearchAccess, expiry int) *Aiven {
	aiven := Aiven{
		context.Background(),
		innClient,
		Properties{
			Username:   username,
			Namespace:  namespace,
			SecretName: secretName,
			Expiry:     expiry,
			Service:    service,
		},
	}

	service.Setup(&services2.ServiceSetup{
		Instance: instance,
		Pool:     pool,
		Access:   access,
	})

	return &aiven
}

func (a *Aiven) GenerateApplication() (*aiven_nais_io_v1.AivenApplication, error) {
	properties := a.Properties
	namespace := v1.Namespace{}

	err := common.ValidateNamespace(a.Ctx, a.Client, properties.Namespace, &namespace)
	if err != nil {
		return nil, err
	}
	properties.Namespace = namespace.Name

	secretName, err := common.SetSecretName(properties.SecretName, properties.Username, properties.Namespace)
	if err != nil {
		return nil, err
	}

	aivenApp := *a.AivenApplication(secretName)
	err = a.CreateOrUpdate(&aivenApp)
	if err != nil {
		return nil, fmt.Errorf("create/update: %v", err)
	}
	return &aivenApp, nil
}

func (a Aiven) AivenApplication(secretName string) *aiven_nais_io_v1.AivenApplication {
	name := strings.ReplaceAll(a.Properties.Username, ".", "-")
	expiresAt := time.Now().AddDate(0, 0, a.Properties.Expiry)
	applicationSpec := aiven_nais_io_v1.AivenApplicationSpec{
		SecretName: secretName,
		Protected:  DefaultProtected,
		ExpiresAt: &metav1.Time{
			Time: expiresAt,
		},
	}

	a.Properties.Service.Apply(&applicationSpec, a.Properties.Namespace)

	app := aiven_nais_io_v1.NewAivenApplicationBuilder(name, a.Properties.Namespace).WithSpec(applicationSpec).Build()
	return &app
}

func (a Aiven) getExisting(existingAivenApp *aiven_nais_io_v1.AivenApplication) error {
	return a.Client.Get(a.Ctx, ctrl.ObjectKey{
		Namespace: a.Properties.Namespace,
		Name:      a.Properties.Username,
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
			log.Default().Printf("AivenApplication: '%v' created.", aivenApp.Name)
		}
	} else {
		if len(existingAivenApp.GetObjectMeta().GetOwnerReferences()) > 0 {
			return fmt.Errorf("username '%s' is owned by another resource; overwrite is not allowed", a.Properties.Username)
		}

		aivenApp.SetResourceVersion(existingAivenApp.GetResourceVersion())
		err = a.Client.Update(a.Ctx, aivenApp)
		if err != nil {
			return err
		}
		log.Default().Printf("AivenApplication: '%v' updated.", aivenApp.Name)
	}
	return nil
}
