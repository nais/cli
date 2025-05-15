package config_test

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/migrate/config"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	nais_io_v1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	liberatorscheme "github.com/nais/liberator/pkg/scheme"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func fakeClient(t *testing.T, objects ...ctrl.Object) ctrl.Client {
	t.Helper()
	scheme, err := liberatorscheme.All()
	if err != nil {
		t.Fatalf("failed to create scheme: %v", err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func TestConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("instance config", func(t *testing.T) {
		t.Run("resolve when app is not found", func(t *testing.T) {
			iCfg := &config.InstanceConfig{}
			client := fakeClient(t)

			t.Run("it returns an error", func(t *testing.T) {
				err := iCfg.Resolve(ctx, client, appName, namespace)
				if err == nil {
					t.Error("expected error, got nil")
				}
			})
		})

		t.Run("resolve when app is found", func(t *testing.T) {
			t.Run("without sqlinstances", func(t *testing.T) {
				client := fakeClient(t, &nais_io_v1alpha1.Application{
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
				})

				t.Run("returns correct error", func(t *testing.T) {
					iCfg := &config.InstanceConfig{}
					err := iCfg.Resolve(ctx, client, appName, namespace)
					if !errors.Is(err, config.ErrMissingSqlInstance) {
						t.Errorf("expected ErrMissingSqlInstance, got: %v", err)
					}
				})
			})

			t.Run("with sqlinstances", func(t *testing.T) {
				client := fakeClient(t, &nais_io_v1alpha1.Application{
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
				for _, tc := range []struct {
					description string
					source      tableEntry
					target      tableEntry
				}{
					{description: "resolves config from app", source: tableEntry{}, target: initialEntry},
					{description: "resolves config from passed config", source: passedEntry, target: passedEntry},
				} {
					t.Run(tc.description, func(t *testing.T) {
						iCfg := &config.InstanceConfig{}
						tc.source.Apply(iCfg)
						err := iCfg.Resolve(ctx, client, appName, namespace)
						if err != nil {
							t.Errorf("unexpected error: %v", err)
						}
						if iCfg.InstanceName != tc.target.InstanceName {
							t.Errorf("expected InstanceName %v, got %v", tc.target.InstanceName, iCfg.InstanceName)
						}
						if iCfg.Tier != tc.target.Tier {
							t.Errorf("expected Tier %v, got %v", tc.target.Tier, iCfg.Tier)
						}
						if iCfg.DiskSize != tc.target.DiskSize {
							t.Errorf("expected DiskSize %v, got %v", tc.target.DiskSize, iCfg.DiskSize)
						}
						if iCfg.DiskAutoresize != tc.target.DiskAutoresize {
							t.Errorf("expected DiskAutoresize %v, got %v", tc.target.DiskAutoresize, iCfg.DiskAutoresize)
						}
						if iCfg.Type != tc.target.Type {
							t.Errorf("expected Type %v, got %v", tc.target.Type, iCfg.Type)
						}
					})
				}
			})

			t.Run("with sqlinstance without optional values", func(t *testing.T) {
				client := fakeClient(t, &nais_io_v1alpha1.Application{
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
				for _, tc := range []struct {
					description string
					source      tableEntry
					target      tableEntry
				}{
					{description: "resolve config from app", source: tableEntry{}, target: initialEntry},
					{description: "resolves config from passed config", source: passedEntry, target: passedEntry},
				} {
					t.Run(tc.description, func(t *testing.T) {
						iCfg := &config.InstanceConfig{}
						tc.source.Apply(iCfg)
						err := iCfg.Resolve(ctx, client, appName, namespace)
						if err != nil {
							t.Errorf("unexpected error: %v", err)
						}
						if iCfg.InstanceName != tc.target.InstanceName {
							t.Errorf("expected InstanceName %v, got %v", tc.target.InstanceName, iCfg.InstanceName)
						}
						if iCfg.Tier != tc.target.Tier {
							t.Errorf("expected Tier %v, got %v", tc.target.Tier, iCfg.Tier)
						}
						if iCfg.DiskSize != tc.target.DiskSize {
							t.Errorf("expected DiskSize %v, got %v", tc.target.DiskSize, iCfg.DiskSize)
						}
						if iCfg.DiskAutoresize != tc.target.DiskAutoresize {
							t.Errorf("expected DiskAutoresize %v, got %v", tc.target.DiskAutoresize, iCfg.DiskAutoresize)
						}
						if iCfg.Type != tc.target.Type {
							t.Errorf("expected Type %v, got %v", tc.target.Type, iCfg.Type)
						}
					})
				}
			})
		})

		t.Run("populate from config map", func(t *testing.T) {
			t.Run("it populates the instance config", func(t *testing.T) {
				iCfg := &config.InstanceConfig{}
				configMap := &corev1.ConfigMap{
					Data: map[string]string{
						"PREFIX_INSTANCE_NAME":            initialInstanceName,
						"PREFIX_INSTANCE_TIER":            initialInstanceTier,
						"PREFIX_INSTANCE_DISKSIZE":        strconv.Itoa(initialDiskSize),
						"PREFIX_INSTANCE_DISK_AUTORESIZE": strconv.FormatBool(initialAutoresize),
						"PREFIX_INSTANCE_TYPE":            string(initialInstanceType),
					},
				}
				iCfg.PopulateFromConfigMap(configMap, "PREFIX")
				if iCfg.InstanceName != option.Some(initialInstanceName) {
					t.Errorf("expected InstanceName %v, got %v", initialInstanceName, iCfg.InstanceName)
				}
				if iCfg.Tier != option.Some(initialInstanceTier) {
					t.Errorf("expected Tier %v, got %v", initialInstanceTier, iCfg.Tier)
				}
				if iCfg.DiskSize != option.Some(initialDiskSize) {
					t.Errorf("expected DiskSize %v, got %v", initialDiskSize, iCfg.DiskSize)
				}
				if iCfg.DiskAutoresize != option.Some(initialAutoresize) {
					t.Errorf("expected DiskAutoresize %v, got %v", initialAutoresize, iCfg.DiskAutoresize)
				}
				if iCfg.Type != option.Some(string(initialInstanceType)) {
					t.Errorf("expected Type %v, got %v", initialInstanceType, iCfg.Type)
				}
			})
		})
	})

	t.Run("test Config", func(t *testing.T) {
		t.Run("migration name", func(t *testing.T) {
			getConfig := func() config.Config {
				return config.Config{
					AppName:   "some-app",
					Namespace: "test-namespace",
					Target: config.InstanceConfig{
						InstanceName: option.Some("target-instance"),
					},
				}
			}

			t.Run("generates valid mgiration name", func(t *testing.T) {
				verify := func(t *testing.T, cfg config.Config, expected string) {
					t.Helper()
					actual := cfg.MigrationName()
					if len(actual) > 63 {
						t.Errorf("expected length <= 63, got %d", len(actual))
					}
					if actual != expected {
						t.Errorf("expected %s, got %s", expected, actual)
					}
				}

				t.Run("happy path with reasonable lengths for app and instance", func(t *testing.T) {
					cfg := getConfig()
					verify(t, cfg, "migration-some-app-target-instance")
				})

				t.Run("very long app name", func(t *testing.T) {
					cfg := getConfig()
					cfg.AppName = "some-unnecessarily-long-app-name-that-should-be-truncated"
					verify(t, cfg, "migration-some-unnecessarily-long-app-name-that-should-377bba1c")
				})

				t.Run("very long instance name", func(t *testing.T) {
					cfg := getConfig()
					cfg.Target.InstanceName = option.Some("some-unnecessarily-long-instance-name-that-should-be-truncated")
					verify(t, cfg, "migration-some-app-some-unnecessarily-long-instance-na-59326cd8")
				})
			})
		})
	})
}
