package cloudsql

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			secret := &corev1.Secret{
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

			k8sClient := fake.NewClientset(secret)
			dbInfo := &CloudSQLDBInfo{DBInfo: &DBInfo{
				k8sClient: k8sClient,
				namespace: namespace,
				appName:   appName,
			}}
			dbConnectionInfo := &ConnectionInfo{
				username: string(secret.Data["DB_USERNAME"]),
				dbName:   string(secret.Data["DB_DATABASE"]),
				password: string(secret.Data["DB_PASSWORD"]),
				url:      parseTestURL(t, secret.Data["DB_URL"]),
				jdbcUrl:  parseTestURL(t, secret.Data["DB_JDBC_URL"]),
			}
			dbConnectionInfo.SetPassword(newPassword)

			if err := updateKubernetesSecret(t.Context(), dbInfo, dbConnectionInfo); err != nil {
				t.Fatalf("failed to update Kubernetes secret: %v", err)
			}

			actual, err := k8sClient.CoreV1().Secrets(secret.Namespace).Get(t.Context(), secret.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("failed to get secret: %v", err)
			}

			for _, assert := range tc.assertSecret {
				assert(t, actual)
			}
		})
	}
}

type SecretPrep func(secret *corev1.Secret)

func AddPassword(secret *corev1.Secret) {
	secret.Data["DB_PASSWORD"] = []byte(oldPassword)
}

func AddUrl(secret *corev1.Secret) {
	secret.Data["DB_URL"] = fmt.Appendf(nil, pgUrlTmpl, oldPassword)
}

func AddJdbcUrl(secret *corev1.Secret) {
	secret.Data["DB_JDBC_URL"] = fmt.Appendf(nil, jdbcUrlTmpl, oldPassword)
}

type AssertSecret func(t *testing.T, actual *corev1.Secret)

func parseTestURL(t *testing.T, value []byte) *url.URL {
	t.Helper()

	if len(value) == 0 {
		return nil
	}
	u, err := url.Parse(string(value))
	if err != nil {
		t.Fatalf("failed to parse test URL: %v", err)
	}
	return u
}

func assertSecretURL(t *testing.T, secret *corev1.Secret, key string, want *url.URL) {
	t.Helper()

	got := parseTestURL(t, secret.Data[key])
	if got == nil {
		t.Fatalf("expected %s to be set", key)
	}
	gotBase, _, _ := strings.Cut(got.String(), "?")
	wantBase, _, _ := strings.Cut(want.String(), "?")
	if gotBase != wantBase {
		t.Errorf("%s URL without query: got %q, want %q", key, gotBase, wantBase)
	}
	if gotQuery, wantQuery := got.Query().Encode(), want.Query().Encode(); gotQuery != wantQuery {
		t.Errorf("%s query: got %q, want %q", key, gotQuery, wantQuery)
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
	assertSecretURL(t, actual, "DB_URL", newPgUrl)
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
	assertSecretURL(t, actual, "DB_JDBC_URL", newJdbcUrl)
}

func HasNoJdbcUrl(t *testing.T, actual *corev1.Secret) {
	t.Helper()
	_, ok := actual.Data["DB_JDBC_URL"]
	if ok {
		t.Fatalf("expected DB_JDBC_URL to not be set, but it is")
	}
}
