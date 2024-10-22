package config

import (
	"context"
	"fmt"
	"strconv"

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
	cfgMap    *corev1.ConfigMap
}

type InstanceConfig struct {
	InstanceName   option.Option[string]
	Tier           option.Option[string]
	DiskAutoresize option.Option[bool]
	DiskSize       option.Option[int]
	Type           option.Option[string]
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

	ic.DiskAutoresize = ic.DiskAutoresize.OrMaybe(func() option.Option[bool] {
		autoresize := app.Spec.GCP.SqlInstances[0].DiskAutoresize
		if autoresize {
			return option.Some(true)
		}
		return option.None[bool]()
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

func (c *Config) MigrationName() string {
	return fmt.Sprintf("migration-%s-%s", c.AppName, c.Target.InstanceName)
}

func (c *Config) CreateConfigMap() *corev1.ConfigMap {
	data := map[string]string{
		"APP_NAME":  c.AppName,
		"NAMESPACE": c.Namespace,
	}

	c.Target.InstanceName.Do(dataBuilder[string](data, "TARGET_INSTANCE_NAME"))
	c.Target.Tier.Do(dataBuilder[string](data, "TARGET_INSTANCE_TIER"))
	c.Target.DiskAutoresize.Do(dataBuilder[bool](data, "TARGET_INSTANCE_DISK_AUTORESIZE"))
	c.Target.DiskSize.Do(dataBuilder[int](data, "TARGET_INSTANCE_DISKSIZE"))
	c.Target.Type.Do(dataBuilder[string](data, "TARGET_INSTANCE_TYPE"))

	c.Source.InstanceName.Do(dataBuilder[string](data, "SOURCE_INSTANCE_NAME"))
	c.Source.Tier.Do(dataBuilder[string](data, "SOURCE_INSTANCE_TIER"))
	c.Source.DiskAutoresize.Do(dataBuilder[bool](data, "SOURCE_INSTANCE_DISK_AUTORESIZE"))
	c.Source.DiskSize.Do(dataBuilder[int](data, "SOURCE_INSTANCE_DISKSIZE"))
	c.Source.Type.Do(dataBuilder[string](data, "SOURCE_INSTANCE_TYPE"))

	c.cfgMap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.MigrationName(),
			Namespace: c.Namespace,
			Labels: map[string]string{
				"migrator.nais.io/migration-name":       c.MigrationName(),
				"migrator.nais.io/app-name":             c.AppName,
				"migrator.nais.io/target-instance-name": c.Target.InstanceName.String(),
			},
			Annotations: map[string]string{
				"migrator.nais.io/created-by": "nais/cli",
			},
		},
		Data: data,
	}
	return c.cfgMap
}

func (c *Config) PopulateFromConfigMap(ctx context.Context, client ctrl.Client) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := client.Get(ctx, ctrl.ObjectKey{Namespace: c.Namespace, Name: c.MigrationName()}, configMap)
	if err != nil {
		return nil, err
	}

	c.Source.InstanceName = option.Some(configMap.Data["SOURCE_INSTANCE_NAME"])
	c.Source.Tier = c.Source.Tier.OrMaybe(func() option.Option[string] {
		sourceTier, ok := configMap.Data["SOURCE_INSTANCE_TIER"]
		if !ok {
			return option.None[string]()
		}
		return option.Some(sourceTier)
	})
	c.Source.DiskAutoresize = c.Source.DiskAutoresize.OrMaybe(func() option.Option[bool] {
		sourceAutoresize, ok := configMap.Data["SOURCE_INSTANCE_DISK_AUTORESIZE"]
		if !ok {
			return option.None[bool]()
		}

		autoresize, err := strconv.ParseBool(sourceAutoresize)
		if err != nil {
			panic("BUG: converting source disk autoresize: " + err.Error())
		}
		return option.Some(autoresize)
	})
	c.Source.DiskSize = c.Source.DiskSize.OrMaybe(func() option.Option[int] {
		sourceDiskSize, ok := configMap.Data["SOURCE_INSTANCE_DISKSIZE"]
		if !ok {
			return option.None[int]()
		}

		diskSize, err := strconv.Atoi(sourceDiskSize)
		if err != nil {
			panic("BUG: converting source disk size: " + err.Error())
		}
		return option.Some(diskSize)
	})
	c.Source.Type = c.Source.Type.OrMaybe(func() option.Option[string] {
		sourceType, ok := configMap.Data["SOURCE_INSTANCE_TYPE"]
		if !ok {
			return option.None[string]()
		}
		return option.Some(sourceType)
	})

	c.Target.InstanceName = option.Some(configMap.Data["TARGET_INSTANCE_NAME"])
	c.Target.Tier = c.Target.Tier.OrMaybe(func() option.Option[string] {
		targetTier, ok := configMap.Data["TARGET_INSTANCE_TIER"]
		if !ok {
			return option.None[string]()
		}
		return option.Some(targetTier)
	})
	c.Target.DiskAutoresize = c.Target.DiskAutoresize.OrMaybe(func() option.Option[bool] {
		targetAutoresize, ok := configMap.Data["TARGET_INSTANCE_DISK_AUTORESIZE"]
		if !ok {
			return option.None[bool]()
		}

		autoresize, err := strconv.ParseBool(targetAutoresize)
		if err != nil {
			panic("BUG: converting source disk autoresize: " + err.Error())
		}
		return option.Some(autoresize)
	})
	c.Target.DiskSize = c.Target.DiskSize.OrMaybe(func() option.Option[int] {
		targetDiskSize, ok := configMap.Data["TARGET_INSTANCE_DISKSIZE"]
		if !ok {
			return option.None[int]()
		}

		diskSize, err := strconv.Atoi(targetDiskSize)
		if err != nil {
			panic("BUG: converting target disk size: " + err.Error())
		}
		return option.Some(diskSize)
	})
	c.Target.Type = c.Target.Type.OrMaybe(func() option.Option[string] {
		targetType, ok := configMap.Data["TARGET_INSTANCE_TYPE"]
		if !ok {
			return option.None[string]()
		}
		return option.Some(targetType)
	})

	c.cfgMap = configMap
	return c.cfgMap, nil
}

func (c *Config) GetConfigMap() *corev1.ConfigMap {
	if c.cfgMap == nil {
		panic("BUG: ConfigMap not initialized")
	}
	return c.cfgMap
}

func dataBuilder[T any](data map[string]string, key string) func(T) {
	return func(v T) {
		data[key] = fmt.Sprintf("%v", v)
	}
}
