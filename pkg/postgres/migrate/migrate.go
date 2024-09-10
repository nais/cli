package migrate

import (
	"context"
	"fmt"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

type Migrator struct {
	client ctrl.Client
	cfg    Config
}

func NewMigrator(client ctrl.Client, cfg Config) *Migrator {
	return &Migrator{
		client,
		cfg,
	}
}

func (m *Migrator) Setup(ctx context.Context) error {
	fmt.Println("Resolving target instance config")
	err := m.cfg.Target.Resolve(ctx, m.client, m.cfg.AppName, m.cfg.Namespace)
	if err != nil {
		return err
	}
	fmt.Printf("Resolved target:\n%s\n", m.cfg.Target.String())

	fmt.Println("Resolving source instance config")
	err = m.cfg.Source.Resolve(ctx, m.client, m.cfg.AppName, m.cfg.Namespace)
	if err != nil {
		return err
	}
	fmt.Printf("Resolved source:\n%s\n", m.cfg.Source.String())

	fmt.Println("Creating ConfigMap")
	cfgMap := m.cfg.CreateConfigMap()
	err = m.client.Create(ctx, &cfgMap)
	if err != nil {
		return fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	fmt.Println("TODO: Do more stuff")
	return fmt.Errorf("TODO: Do more stuff")
}
