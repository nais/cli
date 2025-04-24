package postgres

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	core_v1 "k8s.io/api/core/v1"
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

var _ = Describe("Password", func() {
	var k8sClient *fake.Clientset
	var secret *core_v1.Secret

	BeforeEach(func() {
		k8sClient = fake.NewSimpleClientset()
		secret = &core_v1.Secret{
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
	})

	DescribeTableSubtree("",
		func(test test) {
			var dbInfo *DBInfo
			var dbConnectionInfo *ConnectionInfo

			BeforeEach(func() {
				for _, prep := range test.secretPrep {
					prep(secret)
				}

				err := k8sClient.Tracker().Add(secret)
				Expect(err).To(BeNil())

				dbInfo = createDbInfo(k8sClient)
				dbConnectionInfo = createConnectionInfo(*secret, dbInfo.connectionName)
			})

			It("rotating password", func(ctx context.Context) {
				dbConnectionInfo.SetPassword(newPassword)

				err := updateKubernetesSecret(ctx, dbInfo, dbConnectionInfo)
				Expect(err).To(BeNil())

				gvr, _ := meta.UnsafeGuessKindToResource(secret.GroupVersionKind())
				actual, err := k8sClient.Tracker().Get(gvr, secret.Namespace, secret.Name)
				Expect(err).To(BeNil())

				actualSecret, ok := actual.(*core_v1.Secret)
				Expect(ok).To(BeTrue())

				for _, assert := range test.assertSecret {
					assert(actualSecret)
				}
			})
		},
		Entry("has only password", test{
			secretPrep:   []SecretPrep{AddPassword},
			assertSecret: []AssertSecret{HasPassword, HasNoUrl, HasNoJdbcUrl},
		}),
		Entry("has password and url", test{
			secretPrep:   []SecretPrep{AddPassword, AddUrl},
			assertSecret: []AssertSecret{HasPassword, HasUrl, HasJdbcUrl},
		}),
		Entry("has all", test{
			secretPrep:   []SecretPrep{AddPassword, AddUrl, AddJdbcUrl},
			assertSecret: []AssertSecret{HasPassword, HasUrl, HasJdbcUrl},
		}),
		Entry("has password and jdbc url", test{
			secretPrep:   []SecretPrep{AddPassword, AddJdbcUrl},
			assertSecret: []AssertSecret{HasPassword, HasNoUrl, HasJdbcUrl},
		}),
	)
})

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

type SecretPrep func(secret *core_v1.Secret)

func AddPassword(secret *core_v1.Secret) {
	secret.Data["DB_PASSWORD"] = []byte(oldPassword)
}

func AddUrl(secret *core_v1.Secret) {
	secret.Data["DB_URL"] = []byte(fmt.Sprintf(pgUrlTmpl, oldPassword))
}

func AddJdbcUrl(secret *core_v1.Secret) {
	secret.Data["DB_JDBC_URL"] = []byte(fmt.Sprintf(jdbcUrlTmpl, oldPassword))
}

type AssertSecret func(actual *core_v1.Secret)

func EqualUrlNoQuery(expected *url.URL) types.GomegaMatcher {
	expectedNoQuery, _, _ := strings.Cut(expected.String(), "?")
	return WithTransform(func(actual *url.URL) string {
		actualNoQuery, _, _ := strings.Cut(actual.String(), "?")
		return actualNoQuery
	}, Equal(expectedNoQuery))
}

func EqualQuery(expected *url.URL) types.GomegaMatcher {
	expectedQuery := expected.Query()
	return WithTransform(func(actual *url.URL) url.Values {
		return actual.Query()
	}, Equal(expectedQuery))
}

func HasPassword(actual *core_v1.Secret) {
	By("should have new password in DB_PASSWORD", func() {
		Expect(actual.Data["DB_PASSWORD"]).To(Equal([]byte(newPassword)))
	})
}

func HasUrl(actual *core_v1.Secret) {
	By("should have new password in DB_URL", func() {
		u, err := url.Parse(string(actual.Data["DB_URL"]))
		Expect(err).To(BeNil())
		Expect(u).To(SatisfyAll(
			EqualUrlNoQuery(newPgUrl),
			EqualQuery(newPgUrl),
		))
	})
}

func HasNoUrl(actual *core_v1.Secret) {
	By("should not have DB_URL", func() {
		_, ok := actual.Data["DB_URL"]
		Expect(ok).To(BeFalse())
	})
}

func HasJdbcUrl(actual *core_v1.Secret) {
	By("should have new password in DB_JDBC_URL", func() {
		u, err := url.Parse(string(actual.Data["DB_JDBC_URL"]))
		Expect(err).To(BeNil())
		Expect(u).To(SatisfyAll(
			EqualUrlNoQuery(newJdbcUrl),
			EqualQuery(newJdbcUrl),
		))
	})
}

func HasNoJdbcUrl(actual *core_v1.Secret) {
	By("should not have DB_JDBC_URL", func() {
		_, ok := actual.Data["DB_JDBC_URL"]
		Expect(ok).To(BeFalse())
	})
}
