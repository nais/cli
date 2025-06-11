package config

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/command/flag"
	nais_io_v1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	"github.com/nais/liberator/pkg/namegen"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	AppName   string
	Namespace flag.Namespace
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

var ErrMissingSqlInstance = errors.New("ErrMissingSqlInstance")

func (ic *InstanceConfig) String() string {
	return fmt.Sprintf("Name: %v\nTier: %v\nDiskSize: %v\nType: %v\n", ic.InstanceName, ic.Tier, ic.DiskSize, ic.Type)
}

func (ic *InstanceConfig) Resolve(ctx context.Context, client ctrl.Client, appName string, namespace flag.Namespace) error {
	app := &nais_io_v1alpha1.Application{}
	err := client.Get(ctx, ctrl.ObjectKey{Namespace: string(namespace), Name: appName}, app)
	if err != nil {
		return err
	}

	if app.Spec.GCP == nil || len(app.Spec.GCP.SqlInstances) == 0 {
		return fmt.Errorf("no sql instances found in app spec, %w", ErrMissingSqlInstance)
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

func makeKey(prefix, key string) string {
	return fmt.Sprintf("%s_%s", prefix, key)
}

func (ic *InstanceConfig) PopulateFromConfigMap(configMap *corev1.ConfigMap, prefix string) {
	ic.InstanceName = option.Some(configMap.Data[makeKey(prefix, "INSTANCE_NAME")])
	ic.Tier = ic.Tier.OrMaybe(func() option.Option[string] {
		configTier, ok := configMap.Data[makeKey(prefix, "INSTANCE_TIER")]
		if !ok {
			return option.None[string]()
		}
		return option.Some(configTier)
	})
	ic.DiskAutoresize = ic.DiskAutoresize.OrMaybe(func() option.Option[bool] {
		configAutoresize, ok := configMap.Data[makeKey(prefix, "INSTANCE_DISK_AUTORESIZE")]
		if !ok {
			return option.None[bool]()
		}

		autoresize, err := strconv.ParseBool(configAutoresize)
		if err != nil {
			panic(fmt.Sprintf("BUG: converting %s disk autoresize: %v", prefix, err.Error()))
		}
		return option.Some(autoresize)
	})
	ic.DiskSize = ic.DiskSize.OrMaybe(func() option.Option[int] {
		configDiskSize, ok := configMap.Data[makeKey(prefix, "INSTANCE_DISKSIZE")]
		if !ok {
			return option.None[int]()
		}

		diskSize, err := strconv.Atoi(configDiskSize)
		if err != nil {
			panic(fmt.Sprintf("BUG: converting %s disk size: %v", prefix, err.Error()))
		}
		return option.Some(diskSize)
	})
	ic.Type = ic.Type.OrMaybe(func() option.Option[string] {
		configType, ok := configMap.Data[makeKey(prefix, "INSTANCE_TYPE")]
		if !ok {
			return option.None[string]()
		}
		return option.Some(configType)
	})
}

func (c *Config) MigrationName() string {
	name := fmt.Sprintf("migration-%s-%s", c.AppName, c.Target.InstanceName)
	maxlen := validation.DNS1123LabelMaxLength

	if len(name) > maxlen {
		truncated, err := namegen.ShortName(name, maxlen)
		if err != nil {
			panic(fmt.Sprintf("BUG: generating migration name: %v", err.Error()))
		}
		return truncated
	}

	return name
}

func (c *Config) CreateConfigMap() *corev1.ConfigMap {
	data := map[string]string{
		"APP_NAME":  c.AppName,
		"NAMESPACE": string(c.Namespace),
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
			Namespace: string(c.Namespace),
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
	err := client.Get(ctx, ctrl.ObjectKey{Namespace: string(c.Namespace), Name: c.MigrationName()}, configMap)
	if err != nil {
		return nil, err
	}

	c.Source.PopulateFromConfigMap(configMap, "SOURCE")
	c.Target.PopulateFromConfigMap(configMap, "TARGET")

	c.cfgMap = configMap
	return c.cfgMap, nil
}

func dataBuilder[T any](data map[string]string, key string) func(T) {
	return func(v T) {
		data[key] = fmt.Sprintf("%v", v)
	}
}
