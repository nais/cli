package aiven

import (
	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	"github.com/nais/nais-d/pkg/client"
	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

var scheme = runtime.NewScheme()

func TestAiven_GenerateApplication(t *testing.T) {

	client.InitScheme(scheme)

	username := "user"
	team := "team"
	pool := "pool"
	secretName := "secret-name"
	expiry := 1

	t.Run(username, func(t *testing.T) {

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
				Name:      team,
			},
		}
		fakeClient := fake.NewFakeClientWithScheme(scheme, &namespace, &aivenApp)
		aiven := SetupAiven(fakeClient, username, team, pool, secretName, expiry)
		err := aiven.GenerateApplication()
		assert.NilError(t, err)
	})
}
