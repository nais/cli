package migrate

import (
	"context"
	"fmt"
	"github.com/nais/cli/pkg/option"
	"github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	AppName   string
	Namespace string
	Target    InstanceConfig
	Source    InstanceConfig
}

type InstanceConfig struct {
	InstanceName option.Option[string]
	Tier         option.Option[string]
	DiskSize     option.Option[int]
	Type         option.Option[string]
}

func (ic *InstanceConfig) String() string {
	return fmt.Sprintf("Name: %v\nTier: %v\nDiskSize: %v\nType: %v\n", ic.InstanceName, ic.Tier, ic.DiskSize, ic.Type)
}

func (ic *InstanceConfig) Resolve(ctx context.Context, client ctrl.Client, appName, namespace string) error {
	app := &nais_io_v1alpha1.Application{}
	err := client.Get(ctx, ctrl.ObjectKey{Namespace: namespace, Name: appName}, app)
	if err != nil {
		return err
	}

	ic.InstanceName = ic.InstanceName.Or(func() string {
		name := app.Spec.GCP.SqlInstances[0].Name
		if len(name) == 0 {
			name = app.GetName()
		}
		return name
	})

	ic.Tier = ic.Tier.OrMaybe(func() option.Option[string] {
		tier := app.Spec.GCP.SqlInstances[0].Tier
		if len(tier) == 0 {
			return option.None[string]()
		}
		return option.Some(tier)
	})

	ic.DiskSize = ic.DiskSize.OrMaybe(func() option.Option[int] {
		diskSize := app.Spec.GCP.SqlInstances[0].DiskSize
		if diskSize == 0 {
			return option.None[int]()
		}
		return option.Some(diskSize)
	})

	ic.Type = ic.Type.OrMaybe(func() option.Option[string] {
		instanceType := app.Spec.GCP.SqlInstances[0].Type
		if len(instanceType) == 0 {
			return option.None[string]()
		}
		return option.Some(string(instanceType))
	})

	return nil
}

func (c Config) MigrationName() string {
	return fmt.Sprintf("migration-%s-%s", c.AppName, c.Target.InstanceName)
}

func (c Config) CreateConfigMap() *corev1.ConfigMap {
	data := map[string]string{
		"APP_NAME":  c.AppName,
		"NAMESPACE": c.Namespace,
	}

	c.Target.InstanceName.Do(dataBuilder[string](data, "TARGET_INSTANCE_NAME"))
	c.Target.Tier.Do(dataBuilder[string](data, "TARGET_INSTANCE_TIER"))
	c.Target.DiskSize.Do(dataBuilder[int](data, "TARGET_INSTANCE_DISKSIZE"))
	c.Target.Type.Do(dataBuilder[string](data, "TARGET_INSTANCE_TYPE"))

	c.Source.InstanceName.Do(dataBuilder[string](data, "SOURCE_INSTANCE_NAME"))
	c.Source.Tier.Do(dataBuilder[string](data, "SOURCE_INSTANCE_TIER"))
	c.Source.DiskSize.Do(dataBuilder[int](data, "SOURCE_INSTANCE_DISKSIZE"))
	c.Source.Type.Do(dataBuilder[string](data, "SOURCE_INSTANCE_TYPE"))

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.MigrationName(),
			Namespace: c.Namespace,
			Labels: map[string]string{
				"migrator.nais.io/migration-name": c.MigrationName(),
			},
			Annotations: map[string]string{
				"migrator.nais.io/created-by": "nais/cli",
			},
		},
		Data: data,
	}
}

func dataBuilder[T any](data map[string]string, key string) func(T) {
	return func(v T) {
		data[key] = fmt.Sprintf("%v", v)
	}
}
