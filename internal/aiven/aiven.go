package aiven

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/output"
	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	"github.com/nais/liberator/pkg/namegen"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
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
	Service    aiven_services.Service
}

func Setup(
	ctx context.Context,
	innClient ctrl.Client,
	aivenService aiven_services.Service,
	username, namespace, secretName string,
	expiry uint,
	serviceSetup *aiven_services.ServiceSetup,
) *Aiven {
	aiven := Aiven{
		Ctx:    ctx,
		Client: innClient,
		Properties: Properties{
			Username:   username,
			Namespace:  namespace,
			SecretName: secretName,
			Expiry:     int(expiry), // #nosec G115
			Service:    aivenService,
		},
	}
	aivenService.Setup(serviceSetup)

	return &aiven
}

func (a Aiven) GenerateApplication(out output.Output) (*aiven_nais_io_v1.AivenApplication, error) {
	properties := a.Properties

	if err := validateNamespace(a.Ctx, a.Client, properties.Namespace); err != nil {
		return nil, err
	}
	secretName := properties.SecretName

	if secretName == "" {
		var err error
		secretName, err = createSecretName(properties.Username, properties.Namespace)
		if err != nil {
			return nil, err
		}
	}

	aivenApp := *a.aivenApplication(secretName)
	if err := a.createOrUpdate(&aivenApp, out); err != nil {
		return nil, fmt.Errorf("create/update: %v", err)
	}
	return &aivenApp, nil
}

func (a Aiven) aivenApplication(secretName string) *aiven_nais_io_v1.AivenApplication {
	name := strings.ReplaceAll(a.Properties.Username, ".", "-")
	expiresAt := time.Now().AddDate(0, 0, a.Properties.Expiry)
	applicationSpec := aiven_nais_io_v1.AivenApplicationSpec{
		SecretName: secretName,
		Protected:  true,
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

func (a Aiven) createOrUpdate(aivenApp *aiven_nais_io_v1.AivenApplication, out output.Output) error {
	existingAivenApp := aiven_nais_io_v1.AivenApplication{}
	err := a.getExisting(&existingAivenApp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			err = a.Client.Create(a.Ctx, aivenApp)
			if err != nil {
				return err
			}
			out.Printf("AivenApplication: '%v' created.", aivenApp.Name)
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
		out.Printf("AivenApplication: '%v' updated.", aivenApp.Name)
	}
	return nil
}

func validateNamespace(ctx context.Context, client ctrl.Client, name string) error {
	var namespace v1.Namespace
	err := client.Get(ctx, ctrl.ObjectKey{Name: name}, &namespace)
	if err != nil {
		return fmt.Errorf("get namespace: %w", err)
	}

	return nil
}

func createSecretName(name, namespace string) (string, error) {
	baseName := fmt.Sprintf("%s-%s", name, strings.ReplaceAll(namespace, ".", "-"))
	secretName, err := namegen.ShortName(baseName, 64)
	if err != nil {
		return "", fmt.Errorf("could not create secretName: %s", err)
	}
	return secretName, nil
}
