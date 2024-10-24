package migrate_test

import (
	"context"
	"fmt"
	"github.com/nais/cli/pkg/option"
	"github.com/nais/cli/pkg/postgres/migrate/config"
	"github.com/nais/cli/pkg/postgres/migrate/ui"
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

const (
	sourceName     = "source-instance"
	sourceType     = "source-type"
	sourceDiskSize = 15
	sourceTier     = "source-tier"

	targetName     = "target-instance"
	targetType     = "target-type"
	targetDiskSize = 20
	targetTier     = "target-tier"
)

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
		target = config.InstanceConfig{InstanceName: option.Some(targetName)}

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
			noInstanceApp := &nais_io_v1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-instance",
					Namespace: namespace,
				},
				Spec: nais_io_v1alpha1.ApplicationSpec{},
			}
			alreadyMigratingApp := &nais_io_v1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "already-migrating",
					Namespace: namespace,
				},
				Spec: nais_io_v1alpha1.ApplicationSpec{
					GCP: &nais_io_v1.GCP{
						SqlInstances: []nais_io_v1.CloudSqlInstance{{
							Name: targetName,
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
							Name: targetName,
						}},
					},
				},
			}

			clientBuilder.WithObjects(noInstanceApp, alreadyMigratingApp, app, cfgMap)
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

	Context("ConfigureTarget", func() {
		DescribeTableSubtree("when source config has", func(source config.InstanceConfig) {
			BeforeEach(func() {
				cfg.Source = source

				ui.AskForDiskAutoresize = func(sourceDiskAutoresize option.Option[bool]) func() option.Option[bool] {
					return func() option.Option[bool] {
						return sourceDiskAutoresize
					}
				}
				ui.AskForDiskSize = func(sourceDiskSize option.Option[int]) func() option.Option[int] {
					return func() option.Option[int] {
						return sourceDiskSize
					}
				}
				ui.AskForTier = func(sourceTier string) func() option.Option[string] {
					return func() option.Option[string] {
						return option.Some(sourceTier)
					}
				}
				ui.AskForType = func(sourceType string) func() option.Option[string] {
					return func() option.Option[string] {
						return option.Some(sourceType)
					}
				}
			})

			When("instance type", func() {
				It("target type is set", func() {
					cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName), Type: option.Some(targetType)}
					m = migratorBuilder()
					m.ConfigureTarget()
					Expect(cfg.Target.Type).To(Equal(option.Some(targetType)))
				})
				It("target type is not set", func() {
					cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName)}
					m = migratorBuilder()
					m.ConfigureTarget()
					Expect(cfg.Target.Type).To(Equal(option.None[string]()))
				})
			})

			When("instance tier", func() {
				It("target tier is set", func() {
					cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName), Tier: option.Some(targetTier)}
					m = migratorBuilder()
					m.ConfigureTarget()
					Expect(cfg.Target.Tier.String()).To(Equal(targetTier))
				})
				It("target tier is not set", func() {
					cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName)}
					m = migratorBuilder()
					m.ConfigureTarget()
					Expect(cfg.Target.Tier).To(Equal(option.None[string]()))
				})
			})

			When("instance disk size", func() {
				It("target disk size is set", func() {
					cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName), DiskSize: option.Some(targetDiskSize)}
					m = migratorBuilder()
					m.ConfigureTarget()
					Expect(cfg.Target.DiskSize.String()).To(Equal(fmt.Sprintf("%v", targetDiskSize)))
				})
				It("target disk size is not set", func() {
					cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName)}
					m = migratorBuilder()
					m.ConfigureTarget()
					Expect(cfg.Target.DiskSize).To(Equal(option.None[int]()))
				})
			})

			When("instance disk autoresize", func() {
				It("target disk autoresize is set to false and target disk size is set", func() {
					cfg.Target = config.InstanceConfig{
						InstanceName:   option.Some(targetName),
						DiskAutoresize: option.Some(false),
						DiskSize:       option.Some(targetDiskSize),
					}
					m = migratorBuilder()
					m.ConfigureTarget()
					Expect(cfg.Target.DiskAutoresize.String()).To(Equal("false"))
					Expect(cfg.Target.DiskSize).To(Equal(option.Some(targetDiskSize)))
				})

				It("target disk autoresize is set to false and target disk size is not set", func() {
					cfg.Target = config.InstanceConfig{
						InstanceName:   option.Some(targetName),
						DiskAutoresize: option.Some(false),
					}
					m = migratorBuilder()
					m.ConfigureTarget()
					Expect(cfg.Target.DiskAutoresize.String()).To(Equal("false"))
					Expect(cfg.Target.DiskSize).To(Equal(option.None[int]()))
				})

				It("target disk autoresize is set to true and target disk size is set", func() {
					cfg.Target = config.InstanceConfig{
						InstanceName:   option.Some(targetName),
						DiskAutoresize: option.Some(true),
						DiskSize:       option.Some(targetDiskSize),
					}
					m = migratorBuilder()
					m.ConfigureTarget()
					Expect(cfg.Target.DiskAutoresize.String()).To(Equal("true"))
				})

				It("target disk autoresize is set to true and target disk size is not set", func() {
					cfg.Target = config.InstanceConfig{
						InstanceName:   option.Some(targetName),
						DiskAutoresize: option.Some(true),
					}
					m = migratorBuilder()
					m.ConfigureTarget()
					Expect(cfg.Target.DiskAutoresize.String()).To(Equal("true"))
				})

				It("target disk autoresize is not set and target disk size is set", func() {
					cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName), DiskSize: option.Some(targetDiskSize)}
					m = migratorBuilder()
					m.ConfigureTarget()
					Expect(cfg.Target.DiskAutoresize).To(Equal(option.None[bool]()))
					Expect(cfg.Target.DiskSize).To(Equal(option.Some(targetDiskSize)))
				})
				It("target disk autoresize is not set and target disk size is not set", func() {
					cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName)}
					m = migratorBuilder()
					m.ConfigureTarget()
					Expect(cfg.Target.DiskAutoresize).To(Equal(option.None[bool]()))
					Expect(cfg.Target.DiskSize).To(Equal(option.None[int]()))
				})
			})
		},
			Entry("only default values", config.InstanceConfig{InstanceName: option.Some(sourceName)}),
			Entry("all values", config.InstanceConfig{
				InstanceName:   option.Some(sourceName),
				Tier:           option.Some(sourceTier),
				DiskAutoresize: option.Some(true),
				DiskSize:       option.Some(sourceDiskSize),
				Type:           option.Some(sourceType),
			}),
			Entry("all values, no autoresize", config.InstanceConfig{
				InstanceName:   option.Some(sourceName),
				Tier:           option.Some(sourceTier),
				DiskAutoresize: option.None[bool](),
				DiskSize:       option.Some(sourceDiskSize),
				Type:           option.Some(sourceType),
			}),
			Entry("autoresize, no disk size", config.InstanceConfig{
				InstanceName:   option.Some(sourceName),
				Tier:           option.Some(sourceTier),
				DiskAutoresize: option.Some(true),
				DiskSize:       option.None[int](),
				Type:           option.Some(sourceType),
			}),
		)
	})
})
