package migrate_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/nais/cli/internal/postgres/migrate/config"
	"github.com/nais/cli/internal/postgres/migrate/ui"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	nais_io_v1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	liberatorscheme "github.com/nais/liberator/pkg/scheme"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	ctrl_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func TestMigrator_Setup(t *testing.T) {
	test := map[string]struct {
		appName     string
		errContains string
	}{
		"return an error if application is not found": {
			appName:     "no-such-app",
			errContains: "not found in namespace",
		},
		"return an error if application has no sql instance": {
			appName:     "no-instance",
			errContains: "no sql instances found in app spec",
		},
		"return an error if migration config already exists": {
			appName:     "already-migrating",
			errContains: "migration config already exists for this application",
		},
	}

	for name, tc := range test {
		t.Run(name, func(t *testing.T) {
			scheme, err := liberatorscheme.All()
			if err != nil {
				t.Fatalf("failed to create scheme: %v", err)
			}
			clientBuilder := ctrl_fake.NewClientBuilder().WithScheme(scheme)
			clientset := fake.NewClientset()

			cfg := config.Config{
				Namespace: namespace,
				Source:    config.InstanceConfig{},
				Target:    config.InstanceConfig{InstanceName: option.Some(targetName)},
				AppName:   tc.appName,
			}
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

			migrator := migrate.NewMigrator(clientBuilder.Build(), clientset, cfg, true, true)

			err = migrator.Setup(context.Background())
			if tc.errContains != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errContains)
				} else if !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected error to contain %q, got %q", tc.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfigureTarget_instance_type(t *testing.T) {
	tests := map[string]struct {
		instance config.InstanceConfig
	}{
		"only default values": {
			instance: config.InstanceConfig{InstanceName: option.Some(sourceName)},
		},
		"all values, no autoresize": {
			instance: config.InstanceConfig{
				InstanceName:   option.Some(sourceName),
				Tier:           option.Some(sourceTier),
				DiskAutoresize: option.None[bool](),
				DiskSize:       option.Some(sourceDiskSize),
				Type:           option.Some(sourceType),
			},
		},
		"autoresize, no disk size": {
			instance: config.InstanceConfig{
				InstanceName:   option.Some(sourceName),
				Tier:           option.Some(sourceTier),
				DiskAutoresize: option.Some(true),
				DiskSize:       option.None[int](),
				Type:           option.Some(sourceType),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := config.Config{
				Namespace: namespace,
				Source:    tc.instance,
			}
			scheme, err := liberatorscheme.All()
			if err != nil {
				t.Fatalf("failed to create scheme: %v", err)
			}
			clientBuilder := ctrl_fake.NewClientBuilder().WithScheme(scheme)
			clientset := fake.NewClientset()
			migratorBuilder := func() *migrate.Migrator {
				client := clientBuilder.Build()
				return migrate.NewMigrator(client, clientset, cfg, true, true)
			}

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

			t.Run("instance type target type is set", func(t *testing.T) {
				cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName), Type: option.Some(targetType)}
				m := migratorBuilder()
				m.ConfigureTarget()
				if cfg.Target.Type.String() != targetType {
					t.Errorf("expected target type %q, got %q", targetType, cfg.Target.Type.String())
				}
			})

			t.Run("instance type target type is not set", func(t *testing.T) {
				cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName)}
				m := migratorBuilder()
				m.ConfigureTarget()
				if cfg.Target.Type != option.None[string]() {
					t.Errorf("expected target type to be None, got %q", cfg.Target.Type.String())
				}
			})

			t.Run("instance tier target tier is set", func(t *testing.T) {
				cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName), Tier: option.Some(targetTier)}
				m := migratorBuilder()
				m.ConfigureTarget()
				if cfg.Target.Tier.String() != targetTier {
					t.Errorf("expected target tier %q, got %q", targetTier, cfg.Target.Tier.String())
				}
			})
			t.Run("instance tier target tier is not set", func(t *testing.T) {
				cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName)}
				m := migratorBuilder()
				m.ConfigureTarget()
				if cfg.Target.Tier != option.None[string]() {
					t.Errorf("expected target tier to be None, got %q", cfg.Target.Tier.String())
				}
			})
			t.Run("instance disk size target disk size is set", func(t *testing.T) {
				cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName), DiskSize: option.Some(targetDiskSize)}
				m := migratorBuilder()
				m.ConfigureTarget()
				if cfg.Target.DiskSize.String() != fmt.Sprintf("%v", targetDiskSize) {
					t.Errorf("expected target disk size %d, got %s", targetDiskSize, cfg.Target.DiskSize.String())
				}
			})
			t.Run("instance disk size target disk size is not set", func(t *testing.T) {
				cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName)}
				m := migratorBuilder()
				m.ConfigureTarget()
				if cfg.Target.DiskSize != option.None[int]() {
					t.Errorf("expected target disk size to be None, got %s", cfg.Target.DiskSize.String())
				}
			})
			t.Run("instance disk autoresize target disk autoresize is set to false and target disk size is set", func(t *testing.T) {
				cfg.Target = config.InstanceConfig{
					InstanceName:   option.Some(targetName),
					DiskAutoresize: option.Some(false),
					DiskSize:       option.Some(targetDiskSize),
				}
				m := migratorBuilder()
				m.ConfigureTarget()
				if cfg.Target.DiskAutoresize.String() != "false" {
					t.Errorf("expected target disk autoresize to be false, got %s", cfg.Target.DiskAutoresize.String())
				}
				if cfg.Target.DiskSize.String() != fmt.Sprintf("%v", targetDiskSize) {
					t.Errorf("expected target disk size %d, got %s", targetDiskSize, cfg.Target.DiskSize.String())
				}
			})
			t.Run("instance disk autoresize target disk autoresize is set to false and target disk size is not set", func(t *testing.T) {
				cfg.Target = config.InstanceConfig{
					InstanceName:   option.Some(targetName),
					DiskAutoresize: option.Some(false),
				}
				m := migratorBuilder()
				m.ConfigureTarget()
				if cfg.Target.DiskAutoresize.String() != "false" {
					t.Errorf("expected target disk autoresize to be false, got %s", cfg.Target.DiskAutoresize.String())
				}
				if cfg.Target.DiskSize != option.None[int]() {
					t.Errorf("expected target disk size to be None, got %s", cfg.Target.DiskSize.String())
				}
			})
			t.Run("instance disk autoresize target disk autoresize is set to true and target disk size is set", func(t *testing.T) {
				cfg.Target = config.InstanceConfig{
					InstanceName:   option.Some(targetName),
					DiskAutoresize: option.Some(true),
					DiskSize:       option.Some(targetDiskSize),
				}
				m := migratorBuilder()
				m.ConfigureTarget()
				if cfg.Target.DiskAutoresize.String() != "true" {
					t.Errorf("expected target disk autoresize to be true, got %s", cfg.Target.DiskAutoresize.String())
				}
			})
			t.Run("instance disk autoresize target disk autoresize is set to true and target disk size is not set", func(t *testing.T) {
				cfg.Target = config.InstanceConfig{
					InstanceName:   option.Some(targetName),
					DiskAutoresize: option.Some(true),
				}
				m := migratorBuilder()
				m.ConfigureTarget()
				if cfg.Target.DiskAutoresize.String() != "true" {
					t.Errorf("expected target disk autoresize to be true, got %s", cfg.Target.DiskAutoresize.String())
				}
			})
			t.Run("instance disk autoresize target disk autoresize is not set and target disk size is set", func(t *testing.T) {
				cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName), DiskSize: option.Some(targetDiskSize)}
				m := migratorBuilder()
				m.ConfigureTarget()
				if cfg.Target.DiskAutoresize != option.None[bool]() {
					t.Errorf("expected target disk autoresize to be None, got %s", cfg.Target.DiskAutoresize.String())
				}
				if cfg.Target.DiskSize.String() != fmt.Sprintf("%v", targetDiskSize) {
					t.Errorf("expected target disk size %d, got %s", targetDiskSize, cfg.Target.DiskSize.String())
				}
			})
			t.Run("instance disk autoresize target disk autoresize is not set and target disk size is not set", func(t *testing.T) {
				cfg.Target = config.InstanceConfig{InstanceName: option.Some(targetName)}
				m := migratorBuilder()
				m.ConfigureTarget()
				if cfg.Target.DiskAutoresize != option.None[bool]() {
					t.Errorf("expected target disk autoresize to be None, got %s", cfg.Target.DiskAutoresize.String())
				}
				if cfg.Target.DiskSize != option.None[int]() {
					t.Errorf("expected target disk size to be None, got %s", cfg.Target.DiskSize.String())
				}
			})
		})
	}
}
