package aiven

import (
	"context"
	"testing"
	"time"

	"github.com/nais/cli/pkg/cli"
	"github.com/nais/cli/internal/k8s"

	"github.com/nais/cli/internal/aiven/aiven_services"
	aivennaisiov1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	username   = "user"
	team       = "team"
	secretName = "secret-name"
	expiry     = 1
	pool       = "nav-dev"
)

func buildWithScheme(objects ...runtime.Object) *fake.ClientBuilder {
	scheme := runtime.NewScheme()
	k8s.InitScheme(scheme)
	return fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...)
}

func TestGenerateAivenApplicationCreated(t *testing.T) {
	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: team,
		},
	}

	fakeClient := buildWithScheme(&namespace).Build()
	kafka := &aiven_services.Kafka{}
	aiven := Setup(context.Background(), fakeClient, kafka, username, team, secretName, expiry, &aiven_services.ServiceSetup{Pool: pool})
	currentAivenApp, err := aiven.GenerateApplication(cli.Stdout())
	assert.NoError(t, err)

	assert.Equal(t, username, currentAivenApp.Name, "Name has the same value")
	assert.Equal(t, team, currentAivenApp.Namespace, "Namespace has the same value")
	assert.Equal(t, secretName, currentAivenApp.Spec.SecretName, "SecretName has the same value")
	assert.Equal(t, pool, currentAivenApp.Spec.Kafka.Pool, "Pool has the same value")

	assert.True(t, currentAivenApp.Spec.ExpiresAt.After(time.Now()), "Parsed date is still valid")
}

func TestGenerateAivenApplicationUpdated(t *testing.T) {
	aivenApp := aivennaisiov1.AivenApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      username,
			Namespace: team,
		},
	}

	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: team,
		},
	}

	fakeClient := buildWithScheme(&namespace, &aivenApp).Build()
	kafka := &aiven_services.Kafka{}
	aiven := Setup(context.Background(), fakeClient, kafka, username, team, secretName, expiry, &aiven_services.ServiceSetup{Pool: pool})
	currentAivenApp, err := aiven.GenerateApplication(cli.Stdout())
	assert.NoError(t, err)

	assert.Equal(t, username, currentAivenApp.Name, "Name has the same value")
	assert.Equal(t, team, currentAivenApp.Namespace, "Namespace has the same value")
	assert.Equal(t, secretName, currentAivenApp.Spec.SecretName, "SecretName has the same value")
	assert.Equal(t, pool, currentAivenApp.Spec.Kafka.Pool, "Pool has the same value")

	assert.True(t, currentAivenApp.Spec.ExpiresAt.After(time.Now()), "Parsed date is still valid")
}

func TestGenerateAivenApplicationUpdated_HasOwnerReference(t *testing.T) {
	aivenApp := aivennaisiov1.AivenApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      username,
			Namespace: team,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "nais.io/v1alpha1",
					Kind:       "Application",
					Name:       username,
					UID:        "12345",
				},
			},
		},
	}

	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: team,
		},
	}

	fakeClient := buildWithScheme(&namespace, &aivenApp).Build()
	kafka := &aiven_services.Kafka{}
	aiven := Setup(context.Background(), fakeClient, kafka, username, team, secretName, expiry, &aiven_services.ServiceSetup{Pool: pool})
	_, err := aiven.GenerateApplication(cli.Stdout())
	assert.EqualError(t, err, "create/update: username 'user' is owned by another resource; overwrite is not allowed")
}

func TestValidNamespace(t *testing.T) {
	ctx := context.Background()
	namespaceName := "team-namespace"

	namespace := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	fakeClient := buildWithScheme(namespace).Build()
	err := validateNamespace(ctx, fakeClient, namespaceName)
	assert.NoError(t, err)
}
