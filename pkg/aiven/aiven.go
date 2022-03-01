package aiven

import (
	"context"
	"fmt"
	"github.com/nais/cli/pkg/common"
	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"log"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

const (
	DefaultProtected = true
)

type Service int64

const (
	Kafka Service = iota
)

var Services = []string{"kafka"}

func ServiceFromString(service string) (Service, error) {
	switch strings.ToLower(service) {
	case "kafka":
		return Kafka, nil
	default:
		return -1, fmt.Errorf("unknown service: %v", service)
	}
}

func (p Service) String() string {
	return Services[p]
}

type KafkaProperties struct {
	Pool KafkaPool
}

type Aiven struct {
	Ctx        context.Context
	Client     ctrl.Client
	Properties Properties
}

type Properties struct {
	Service    Service
	Username   string
	Namespace  string
	Dest       string
	SecretName string
	Expiry     int
	Kafka      *KafkaProperties
}

func SetupAiven(innClient ctrl.Client, service Service, username, namespace, secretName string, expiry int, pool KafkaPool) *Aiven {
	aiven := &Aiven{
		context.Background(),
		innClient,
		Properties{
			Service:    service,
			Username:   username,
			Namespace:  namespace,
			SecretName: secretName,
			Expiry:     expiry,
		},
	}

	switch service {
	case Kafka:
		aiven.Properties.Kafka = &KafkaProperties{
			Pool: pool,
		}

	}

	return aiven
}

func (a *Aiven) GenerateApplication() (*aiven_nais_io_v1.AivenApplication, error) {
	namespace := v1.Namespace{}
	err := common.ValidateNamespace(a.Ctx, a.Client, a.Properties.Namespace, &namespace)
	if err != nil {
		return nil, err
	}
	a.Properties.Namespace = namespace.Name

	secretName, err := common.SetSecretName(a.Properties.SecretName, a.Properties.Username, a.Properties.Namespace)
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
	applicationSpec := aiven_nais_io_v1.AivenApplicationSpec{
		SecretName: secretName,
		Protected:  DefaultProtected,
		ExpiresAt:  time.Now().AddDate(0, 0, a.Properties.Expiry).Format(time.RFC3339),
	}

	switch a.Properties.Service {
	case Kafka:
		applicationSpec.Kafka = &aiven_nais_io_v1.KafkaSpec{
			Pool: a.Properties.Kafka.Pool.String(),
		}
	}

	name := strings.ReplaceAll(a.Properties.Username, ".", "-")
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
		aivenApp.SetResourceVersion(existingAivenApp.GetResourceVersion())
		err = a.Client.Update(a.Ctx, aivenApp)
		if err != nil {
			return err
		}
		log.Default().Printf("AivenApplication: '%v' updated.", aivenApp.Name)
	}
	return nil
}
