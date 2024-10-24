package migrate_test

import (
	"context"
	"github.com/nais/cli/pkg/option"
	"github.com/nais/cli/pkg/postgres/migrate/config"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	nais_io_v1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	liberatorscheme "github.com/nais/liberator/pkg/scheme"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ctrl_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/nais/cli/pkg/postgres/migrate"
)

const namespace = "test-namespace"

var _ = Describe("Migrator", func() {
	var err error
	var scheme *runtime.Scheme
	var clientBuilder *ctrl_fake.ClientBuilder
	var clientset *fake.Clientset

	var cfg config.Config
	var source config.InstanceConfig
	var target config.InstanceConfig
	var migratorBuilder func() *migrate.Migrator
	var m *migrate.Migrator

	BeforeEach(func() {
		scheme, err = liberatorscheme.All()
		Expect(err).ToNot(HaveOccurred())
		clientBuilder = ctrl_fake.NewClientBuilder().WithScheme(scheme)
		clientset = fake.NewClientset()

		source = config.InstanceConfig{}
		target = config.InstanceConfig{InstanceName: option.Some("target-instance")}

		cfg = config.Config{
			Namespace: namespace,
			Source:    source,
			Target:    target,
		}

		migratorBuilder = func() *migrate.Migrator {
			client := clientBuilder.Build()
			return migrate.NewMigrator(client, clientset, cfg, true, true)
		}
	})

	Context("Setup", func() {
		BeforeEach(func() {
			no_instance_app := &nais_io_v1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-instance",
					Namespace: namespace,
				},
				Spec: nais_io_v1alpha1.ApplicationSpec{},
			}
			already_migrating_app := &nais_io_v1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "already-migrating",
					Namespace: namespace,
				},
				Spec: nais_io_v1alpha1.ApplicationSpec{
					GCP: &nais_io_v1.GCP{
						SqlInstances: []nais_io_v1.CloudSqlInstance{{
							Name: "target-instance",
						}},
					},
				},
			}
			cfgMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "migration-already-exists-config",
					Namespace: namespace,
					Labels:    map[string]string{"migrator.nais.io/app-name": "already-migrating"},
				},
			}
			app := &nais_io_v1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-app",
					Namespace: namespace,
				},
				Spec: nais_io_v1alpha1.ApplicationSpec{
					GCP: &nais_io_v1.GCP{
						SqlInstances: []nais_io_v1.CloudSqlInstance{{
							Name: "target-instance",
						}},
					},
				},
			}

			clientBuilder.WithObjects(no_instance_app, already_migrating_app, app, cfgMap)
		})

		It("should return an error if application is not found", func() {
			cfg.AppName = "no-such-app"
			m = migratorBuilder()
			err := m.Setup(context.Background())
			Expect(err).To(HaveOccurred())
		})

		It("should return an error if application has no sql instance", func() {
			cfg.AppName = "no-instance"
			m = migratorBuilder()
			err := m.Setup(context.Background())
			Expect(err).To(HaveOccurred())
		})

		It("should return an error if migration config already exists", func() {
			cfg.AppName = "already-migrating"
			m = migratorBuilder()
			err := m.Setup(context.Background())
			Expect(err).To(HaveOccurred())
		})
	})
})
