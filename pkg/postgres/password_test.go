package postgres

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
)

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
				"DB_USERNAME": []byte("my-user"),
			},
		}
	})

	It("should rotate password", func(ctx context.Context) {
		secret.Data["DB_PASSWORD"] = []byte("old-password")

		err := k8sClient.Tracker().Add(secret)
		Expect(err).To(BeNil())

		dbInfo := createDbInfo(k8sClient)
		dbConnectionInfo := createDbConnectionInfo()

		err = updateKubernetesSecret(ctx, dbInfo, dbConnectionInfo)
		Expect(err).To(BeNil())

		gvr, _ := meta.UnsafeGuessKindToResource(secret.GroupVersionKind())
		actual, err := k8sClient.Tracker().Get(gvr, secret.Namespace, secret.Name)
		Expect(err).To(BeNil())

		actualSecret, ok := actual.(*core_v1.Secret)
		Expect(ok).To(BeTrue())

		Expect(actualSecret.Data["DB_PASSWORD"]).To(Equal([]byte(newPassword)))
	})
})

func createDbConnectionInfo() *ConnectionInfo {
	return &ConnectionInfo{
		username: "",
		password: newPassword,
		dbName:   "",
		instance: "",
		port:     "",
		host:     "",
		url:      nil,
		jdbcUrl:  nil,
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
		multiDB:        false,
		instanceName:   "my-instance",
		databaseName:   "my-database",
		user:           "my-user",
	}
}
