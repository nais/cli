package aiven

import (
	"github.com/nais/cli/pkg/common"
	"github.com/nais/cli/pkg/test"
	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
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
	aiven := Setup(fakeClient, Kafka, username, team, secretName, "", NavDev, Read, expiry)
	currentAivenApp, err := aiven.GenerateApplication()
	assert.NoError(t, err)

	assert.Equal(t, username, currentAivenApp.Name, "Name has the same value")
	assert.Equal(t, team, currentAivenApp.Namespace, "Namespace has the same value")
	assert.Equal(t, secretName, currentAivenApp.Spec.SecretName, "SecretName has the same value")
	assert.Equal(t, NavDev.String(), currentAivenApp.Spec.Kafka.Pool, "Pool has the same value")

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
	aiven := Setup(fakeClient, Kafka, username, team, secretName, "", NavDev, Read, expiry)
	currentAivenApp, err := aiven.GenerateApplication()
	assert.NoError(t, err)

	assert.Equal(t, username, currentAivenApp.Name, "Name has the same value")
	assert.Equal(t, team, currentAivenApp.Namespace, "Namespace has the same value")
	assert.Equal(t, secretName, currentAivenApp.Spec.SecretName, "SecretName has the same value")
	assert.Equal(t, NavDev.String(), currentAivenApp.Spec.Kafka.Pool, "Pool has the same value")

	assert.True(t, currentAivenApp.Spec.ExpiresAt.After(time.Now()), "Parsed date is still valid")
}

func TestAiven_SetSecretName(t *testing.T) {
	s, err := common.SetSecretName(secretName, username, team)
	assert.NoError(t, err)
	assert.Equal(t, secretName, s, "SecretName has the same value as input")

	s, err = common.SetSecretName("", username, team)
	assert.NoError(t, err)
	assert.Equal(t, "team-user-df60919d", s, "SecretName is generated")
}
