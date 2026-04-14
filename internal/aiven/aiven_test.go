package aiven

import (
	"context"
	"testing"

	"github.com/nais/cli/internal/k8s"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func buildWithScheme(objects ...runtime.Object) *fake.ClientBuilder {
	scheme := runtime.NewScheme()
	k8s.InitScheme(scheme)
	return fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...)
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
