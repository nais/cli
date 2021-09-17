package common

import (
	"context"
	"fmt"
	"github.com/nais/liberator/pkg/namegen"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	FilePermission           = 0775
	MaxServiceUserNameLength = 64
)

func Destination(dest, filename string) string {
	return fmt.Sprintf("%s/%s", dest, filename)
}

func WriteToFile(dest, filename string, value []byte) error {
	err := ioutil.WriteFile(Destination(dest, filename), value, FilePermission)
	if err != nil {
		return err
	}
	return nil
}

func ValidateNamespace(ctx context.Context, client ctrl.Client, name string, namespace *v1.Namespace) error {
	err := client.Get(ctx, ctrl.ObjectKey{
		Name: name,
	}, namespace)
	if err != nil {
		return fmt.Errorf("getting namespace")
	}
	return nil
}

func SetSecretName(secretName, name, namespace string) (string, error) {
	if secretName != "" {
		return secretName, nil
	}
	secretName, err := setSecretName(name, namespace)
	if err != nil {
		return "", fmt.Errorf("could not create secretName: %s", err)
	}
	return secretName, nil
}

func setSecretName(name, namespace string) (string, error) {
	return namegen.ShortName(secretNamePrefix(namespace, strings.ReplaceAll(name, ".", "-")), MaxServiceUserNameLength)
}

func secretNamePrefix(username, team string) string {
	return fmt.Sprintf("%s-%s", team, username)
}
