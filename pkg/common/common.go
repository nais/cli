package common

import (
	"context"
	"fmt"
	"github.com/nais/liberator/pkg/namegen"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	FilePermission           = 0775
	MaxServiceUserNameLength = 64
)

func WriteToFile(dest, filename string, value []byte) error {
	err := ioutil.WriteFile(filepath.Join(dest, filename), value, FilePermission)
	if err != nil {
		return err
	}
	return nil
}

func ValidateNamespace(ctx context.Context, client ctrl.Client, name string, namespace *v1.Namespace) error {
	err := client.Get(ctx, ctrl.ObjectKey{Name: name}, namespace)
	if err != nil {
		return fmt.Errorf("get namespace: %w", err)
	}

	if namespace.GetLabels()["shared"] == "true" {
		return fmt.Errorf("shared namespace is not allowed: %s", name)
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
	return namegen.ShortName(secretNamePrefix(name, strings.ReplaceAll(namespace, ".", "-")), MaxServiceUserNameLength)
}

func secretNamePrefix(username, namespace string) string {
	return fmt.Sprintf("%s-%s", namespace, username)
}

func RequiredSecretDataExists(required map[string]string, secretData map[string][]byte, filetype string) error {
	for key, _ := range required {
		if _, ok := secretData[key]; !ok {
			return fmt.Errorf("can not generate %s config, secret missing required key: %s", filetype, key)
		}
	}
	return nil
}
