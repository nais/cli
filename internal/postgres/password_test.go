package postgres

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	namespace   = "password-ns"
	secretName  = "google-sql-password-app"
	appName     = "password-app"
	newPassword = "new-password"
	oldPassword = "old-password"

	jdbcUrlTmpl = "jdbc:postgresql://localhost:5432/my-database?user=my-user&password=%s"
	pgUrlTmpl   = "postgresql://my-user:%s@localhost:5432/my-database"
)

var (
	newJdbcUrl *url.URL
	newPgUrl   *url.URL
)

func init() {
	var err error
	newJdbcUrl, err = url.Parse(fmt.Sprintf(jdbcUrlTmpl, newPassword))
	if err != nil {
		panic(err)
	}

	newPgUrl, err = url.Parse(fmt.Sprintf(pgUrlTmpl, newPassword))
	if err != nil {
		panic(err)
	}
}

type test struct {
	secretPrep   []SecretPrep
	assertSecret []AssertSecret
}

func TestPassword(t *testing.T) {
	tests := map[string]test{
		"has only password": {
			secretPrep:   []SecretPrep{AddPassword},
			assertSecret: []AssertSecret{HasPassword, HasNoUrl, HasNoJdbcUrl},
		},
		"has password and url": {
			secretPrep:   []SecretPrep{AddPassword, AddUrl},
			assertSecret: []AssertSecret{HasPassword, HasUrl, HasJdbcUrl},
		},
		"has all": {
			secretPrep:   []SecretPrep{AddPassword, AddUrl, AddJdbcUrl},
			assertSecret: []AssertSecret{HasPassword, HasUrl, HasJdbcUrl},
		},
		"has password and jdbc url": {
			secretPrep:   []SecretPrep{AddPassword, AddJdbcUrl},
			assertSecret: []AssertSecret{HasPassword, HasNoUrl, HasJdbcUrl},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			k8sClient := fake.NewClientset()
			secret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"DB_HOST":     []byte("localhost"),
					"DB_PORT":     []byte("5432"),
					"DB_DATABASE": []byte("my-database"),
					"DB_USERNAME": []byte("my-user"),
				},
			}

			for _, prep := range tc.secretPrep {
				prep(secret)
			}

			err := k8sClient.Tracker().Add(secret)
			if err != nil {
				t.Fatalf("failed to add secret to tracker: %v", err)
			}

			dbInfo := createDbInfo(k8sClient)
			dbConnectionInfo := createConnectionInfo(*secret, dbInfo.connectionName)

			dbConnectionInfo.SetPassword(newPassword)

			if err := updateKubernetesSecret(t.Context(), dbInfo, dbConnectionInfo); err != nil {
				t.Fatalf("failed to update Kubernetes secret: %v", err)
			}

			gvr, _ := meta.UnsafeGuessKindToResource(secret.GroupVersionKind())
			actual, err := k8sClient.Tracker().Get(gvr, secret.Namespace, secret.Name)
			if err != nil {
				t.Fatalf("failed to get secret: %v", err)
			}

			actualSecret, ok := actual.(*corev1.Secret)
			if !ok {
				t.Fatalf("expected *corev1.Secret, got %T", actual)
			}

			for _, assert := range tc.assertSecret {
				assert(t, actualSecret)
			}
		})
	}
}

func createDbInfo(k8sClient kubernetes.Interface) *DBInfo {
	return &DBInfo{
		k8sClient:      k8sClient,
		dynamicClient:  nil,
		config:         nil,
		namespace:      namespace,
		appName:        appName,
		projectID:      "project-id",
		connectionName: "connection:name",
	}
}

type SecretPrep func(secret *corev1.Secret)

func AddPassword(secret *corev1.Secret) {
	secret.Data["DB_PASSWORD"] = []byte(oldPassword)
}

func AddUrl(secret *corev1.Secret) {
	secret.Data["DB_URL"] = []byte(fmt.Sprintf(pgUrlTmpl, oldPassword))
}

func AddJdbcUrl(secret *corev1.Secret) {
	secret.Data["DB_JDBC_URL"] = []byte(fmt.Sprintf(jdbcUrlTmpl, oldPassword))
}

type AssertSecret func(t *testing.T, actual *corev1.Secret)

func EqualUrlNoQuery(t *testing.T, expected *url.URL) {
	t.Helper()

	expectedNoQuery, _, _ := strings.Cut(expected.String(), "?")
	actualNoQuery, _, _ := strings.Cut(expected.String(), "?")
	if expectedNoQuery != actualNoQuery {
		t.Fatalf("expected URL without query to be '%s', but got '%s'", expectedNoQuery, actualNoQuery)
	}
	if actualNoQuery != expectedNoQuery {
		t.Fatalf("expected URL without query to be '%s', but got '%s'", expectedNoQuery, actualNoQuery)
	}
}

func EqualQuery(t *testing.T, expected *url.URL) {
	t.Helper()

	expectedQuery := expected.Query()
	actualQuery := expected.Query()
	if actualQuery.Encode() != expectedQuery.Encode() {
		t.Fatalf("expected URL query to be '%s', but got '%s'", expectedQuery.Encode(), actualQuery.Encode())
	}
}

func HasPassword(t *testing.T, actual *corev1.Secret) {
	t.Helper()

	if actual.Data["DB_PASSWORD"] == nil {
		t.Fatalf("expected DB_PASSWORD to be set, but it is not")
	}

	if string(actual.Data["DB_PASSWORD"]) != newPassword {
		t.Fatalf("expected DB_PASSWORD to be '%s', but got '%s'", newPassword, actual.Data["DB_PASSWORD"])
	}
}

func HasUrl(t *testing.T, actual *corev1.Secret) {
	t.Helper()
	u, err := url.Parse(string(actual.Data["DB_URL"]))
	if err != nil {
		t.Fatalf("failed to parse DB_URL: %v", err)
	}
	if u == nil {
		t.Fatalf("DB_URL is nil")
	}
	EqualUrlNoQuery(t, newPgUrl)
	EqualQuery(t, newPgUrl)
}

func HasNoUrl(t *testing.T, actual *corev1.Secret) {
	t.Helper()
	_, ok := actual.Data["DB_URL"]
	if ok {
		t.Fatalf("expected DB_URL to not be set, but it is")
	}
}

func HasJdbcUrl(t *testing.T, actual *corev1.Secret) {
	t.Helper()
	u, err := url.Parse(string(actual.Data["DB_JDBC_URL"]))
	if err != nil {
		t.Fatalf("failed to parse DB_JDBC_URL: %v", err)
	}
	if u == nil {
		t.Fatalf("DB_JDBC_URL is nil")
	}
	EqualUrlNoQuery(t, newJdbcUrl)
	EqualQuery(t, newJdbcUrl)
}

func HasNoJdbcUrl(t *testing.T, actual *corev1.Secret) {
	t.Helper()
	_, ok := actual.Data["DB_JDBC_URL"]
	if ok {
		t.Fatalf("expected DB_JDBC_URL to not be set, but it is")
	}
}
