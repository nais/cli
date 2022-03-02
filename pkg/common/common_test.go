package common

import (
	"context"
	"github.com/nais/cli/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestValidateNamespaceShared(t *testing.T) {
	ctx := context.Background()
	namespaceName := "default"

	namespace := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespaceName,
			Labels: map[string]string{"shared": "true"},
		},
	}

	fakeClient := test.BuildWithScheme(namespace).Build()
	err := ValidateNamespace(ctx, fakeClient, namespaceName, namespace)
	assert.EqualError(t, err, "shared namespace is not allowed: default")
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

	fakeClient := test.BuildWithScheme(namespace).Build()
	err := ValidateNamespace(ctx, fakeClient, namespaceName, namespace)
	assert.NoError(t, err)
}
