package aiven

import (
	"testing"
	"time"

	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nais/cli/pkg/aiven/services"
	"github.com/nais/cli/pkg/common"
	"github.com/nais/cli/pkg/test"
)

const (
	username   = "user"
	team       = "team"
	secretName = "secret-name"
	expiry     = 1
)

func TestAivenGenerateApplicationCreated(t *testing.T) {

	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: team,
		},
	}

	fakeClient := test.BuildWithScheme(&namespace).Build()
	kafka := &services.Kafka{}
	aiven := Setup(fakeClient, kafka, username, team, secretName, "", services.NavDev, services.Read, expiry)
	currentAivenApp, err := aiven.GenerateApplication()
	assert.NoError(t, err)

	assert.Equal(t, username, currentAivenApp.Name, "Name has the same value")
	assert.Equal(t, team, currentAivenApp.Namespace, "Namespace has the same value")
	assert.Equal(t, secretName, currentAivenApp.Spec.SecretName, "SecretName has the same value")
	assert.Equal(t, services.NavDev.String(), currentAivenApp.Spec.Kafka.Pool, "Pool has the same value")

	assert.True(t, currentAivenApp.Spec.ExpiresAt.After(time.Now()), "Parsed date is still valid")
}

func TestAivenGenerateApplicationUpdated(t *testing.T) {
	aivenApp := aiven_nais_io_v1.AivenApplication{
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

	fakeClient := test.BuildWithScheme(&namespace, &aivenApp).Build()
	kafka := &services.Kafka{}
	aiven := Setup(fakeClient, kafka, username, team, secretName, "", services.NavDev, services.Read, expiry)
	currentAivenApp, err := aiven.GenerateApplication()
	assert.NoError(t, err)

	assert.Equal(t, username, currentAivenApp.Name, "Name has the same value")
	assert.Equal(t, team, currentAivenApp.Namespace, "Namespace has the same value")
	assert.Equal(t, secretName, currentAivenApp.Spec.SecretName, "SecretName has the same value")
	assert.Equal(t, services.NavDev.String(), currentAivenApp.Spec.Kafka.Pool, "Pool has the same value")

	assert.True(t, currentAivenApp.Spec.ExpiresAt.After(time.Now()), "Parsed date is still valid")
}

func TestAivenGenerateApplicationUpdated_HasOwnerReference(t *testing.T) {
	aivenApp := aiven_nais_io_v1.AivenApplication{
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

	fakeClient := test.BuildWithScheme(&namespace, &aivenApp).Build()
	kafka := &services.Kafka{}
	aiven := Setup(fakeClient, kafka, username, team, secretName, "", services.NavDev, services.Read, expiry)
	_, err := aiven.GenerateApplication()
	assert.EqualError(t, err, "create/update: username 'user' is owned by another resource; overwrite is not allowed")
}

func TestAiven_SetSecretName(t *testing.T) {
	s, err := common.SetSecretName(secretName, username, team)
	assert.NoError(t, err)
	assert.Equal(t, secretName, s, "SecretName has the same value as input")

	s, err = common.SetSecretName("", username, team)
	assert.NoError(t, err)
	assert.Equal(t, "team-user-df60919d", s, "SecretName is generated")
}
