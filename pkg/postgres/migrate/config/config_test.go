package config_test

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
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"strconv"
)

const (
	appName   = "myapp"
	namespace = "mynamespace"

	initialInstanceName = "myinstance"
	initialInstanceTier = "db-f1-micro"
	initialDiskSize     = 11
	initialAutoresize   = true
	initialInstanceType = nais_io_v1.CloudSqlInstanceTypePostgres15
)

type tableEntry struct {
	InstanceName   option.Option[string]
	Tier           option.Option[string]
	DiskSize       option.Option[int]
	DiskAutoresize option.Option[bool]
	Type           option.Option[string]
}

func (t tableEntry) Apply(iCfg *config.InstanceConfig) {
	t.InstanceName.Do(func(v string) {
		iCfg.InstanceName = option.Some(v)
	})
	t.Tier.Do(func(v string) {
		iCfg.Tier = option.Some(v)
	})
	t.DiskSize.Do(func(v int) {
		iCfg.DiskSize = option.Some(v)
	})
	t.DiskAutoresize.Do(func(v bool) {
		iCfg.DiskAutoresize = option.Some(v)
	})
	t.Type.Do(func(v string) {
		iCfg.Type = option.Some(v)
	})
}

var _ = Describe("config", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("InstanceConfig", func() {
		var iCfg *config.InstanceConfig
		var clientBuilder *fake.ClientBuilder

		BeforeEach(func() {
			scheme, err := liberatorscheme.All()
			Expect(err).ToNot(HaveOccurred())
			clientBuilder = fake.NewClientBuilder().WithScheme(scheme)
			iCfg = &config.InstanceConfig{}
		})

		Context("Resolve", func() {
			When("app is not found", func() {
				It("returns an error", func() {
					client := clientBuilder.Build()
					err := iCfg.Resolve(ctx, client, appName, namespace)
					Expect(err).To(HaveOccurred())
				})
			})

			When("app is found", func() {
				Context("but has no sqlinstances", func() {
					var client ctrl.Client

					BeforeEach(func() {
						client = clientBuilder.WithObjects(&nais_io_v1alpha1.Application{
							TypeMeta: metav1.TypeMeta{
								APIVersion: "nais.io/v1alpha1",
								Kind:       "Application",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name:      appName,
								Namespace: namespace,
							},
							Spec: nais_io_v1alpha1.ApplicationSpec{
								Image: "myimage",
							},
						}).Build()
					})

					It("returns a MissingSqlInstanceError", func() {
						err := iCfg.Resolve(ctx, client, appName, namespace)
						Expect(err).To(HaveOccurred())
					})
				})

				Context("and has sqlinstances", func() {
					var client ctrl.Client

					BeforeEach(func() {
						client = clientBuilder.WithObjects(&nais_io_v1alpha1.Application{
							TypeMeta: metav1.TypeMeta{
								APIVersion: "nais.io/v1alpha1",
								Kind:       "Application",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name:      appName,
								Namespace: namespace,
							},
							Spec: nais_io_v1alpha1.ApplicationSpec{
								Image: "myimage",
								GCP: &nais_io_v1.GCP{
									SqlInstances: []nais_io_v1.CloudSqlInstance{
										{
											Type:           initialInstanceType,
											Name:           initialInstanceName,
											Tier:           initialInstanceTier,
											DiskSize:       initialDiskSize,
											DiskAutoresize: initialAutoresize,
										},
									},
								},
							},
						}).Build()
					})

					initialEntry := tableEntry{
						InstanceName:   option.Some(initialInstanceName),
						Tier:           option.Some(initialInstanceTier),
						DiskSize:       option.Some(initialDiskSize),
						DiskAutoresize: option.Some(initialAutoresize),
						Type:           option.Some(string(initialInstanceType)),
					}
					passedEntry := tableEntry{
						InstanceName:   option.Some("passedName"),
						Tier:           option.Some("passedTier"),
						DiskSize:       option.Some(999),
						DiskAutoresize: option.Some(false),
						Type:           option.Some("passedType"),
					}
					DescribeTable("it correctly resolves config", func(fixture tableEntry, expected tableEntry) {
						fixture.Apply(iCfg)

						err := iCfg.Resolve(ctx, client, appName, namespace)
						Expect(err).ToNot(HaveOccurred())

						Expect(iCfg.InstanceName).To(Equal(expected.InstanceName))
						Expect(iCfg.Tier).To(Equal(expected.Tier))
						Expect(iCfg.Type).To(Equal(expected.Type))
						Expect(iCfg.DiskSize).To(Equal(expected.DiskSize))
						Expect(iCfg.DiskAutoresize).To(Equal(expected.DiskAutoresize))
					},
						Entry("from app", tableEntry{}, initialEntry),
						Entry("from passed config", passedEntry, passedEntry),
					)
				})

				Context("and has sqlinstances without optional values", func() {
					var client ctrl.Client

					BeforeEach(func() {
						client = clientBuilder.WithObjects(&nais_io_v1alpha1.Application{
							TypeMeta: metav1.TypeMeta{
								APIVersion: "nais.io/v1alpha1",
								Kind:       "Application",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name:      appName,
								Namespace: namespace,
							},
							Spec: nais_io_v1alpha1.ApplicationSpec{
								Image: "myimage",
								GCP: &nais_io_v1.GCP{
									SqlInstances: []nais_io_v1.CloudSqlInstance{
										{
											Type: initialInstanceType,
										},
									},
								},
							},
						}).Build()
					})

					initialEntry := tableEntry{
						InstanceName:   option.Some(appName),
						Tier:           option.None[string](),
						DiskSize:       option.None[int](),
						DiskAutoresize: option.None[bool](),
						Type:           option.Some(string(initialInstanceType)),
					}
					passedEntry := tableEntry{
						InstanceName:   option.Some("passedName"),
						Tier:           option.Some("passedTier"),
						DiskSize:       option.Some(999),
						DiskAutoresize: option.Some(false),
						Type:           option.Some("passedType"),
					}
					DescribeTable("it correctly resolves config", func(fixture tableEntry, expected tableEntry) {
						fixture.Apply(iCfg)

						err := iCfg.Resolve(ctx, client, appName, namespace)
						Expect(err).ToNot(HaveOccurred())

						Expect(iCfg.InstanceName).To(Equal(expected.InstanceName))
						Expect(iCfg.Tier).To(Equal(expected.Tier))
						Expect(iCfg.Type).To(Equal(expected.Type))
						Expect(iCfg.DiskSize).To(Equal(expected.DiskSize))
						Expect(iCfg.DiskAutoresize).To(Equal(expected.DiskAutoresize))
					},
						Entry("from app", tableEntry{}, initialEntry),
						Entry("from passed config", passedEntry, passedEntry),
					)
				})
			})
		})

		Context("populateFromConfigMap", func() {
			const prefix = "PREFIX"
			var configMap *corev1.ConfigMap

			BeforeEach(func() {
				configMap = &corev1.ConfigMap{
					Data: map[string]string{
						"PREFIX_INSTANCE_NAME":            initialInstanceName,
						"PREFIX_INSTANCE_TIER":            initialInstanceTier,
						"PREFIX_INSTANCE_DISKSIZE":        strconv.Itoa(initialDiskSize),
						"PREFIX_INSTANCE_DISK_AUTORESIZE": strconv.FormatBool(initialAutoresize),
						"PREFIX_INSTANCE_TYPE":            string(initialInstanceType),
					},
				}
			})

			It("populates the instance config", func() {
				iCfg.PopulateFromConfigMap(configMap, prefix)

				Expect(iCfg.InstanceName).To(Equal(option.Some(initialInstanceName)))
				Expect(iCfg.Tier).To(Equal(option.Some(initialInstanceTier)))
				Expect(iCfg.DiskSize).To(Equal(option.Some(initialDiskSize)))
				Expect(iCfg.DiskAutoresize).To(Equal(option.Some(initialAutoresize)))
				Expect(iCfg.Type).To(Equal(option.Some(string(initialInstanceType))))
			})
		})
	})
})
