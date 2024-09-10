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
	err := m.cfg.Target.Resolve(ctx, m.client, m.cfg.AppName, m.cfg.Namespace)
	if err != nil {
		return err
	}

	err = m.cfg.Source.Resolve(ctx, m.client, m.cfg.AppName, m.cfg.Namespace)
	if err != nil {
		return err
	}

	cfgMap := m.cfg.CreateConfigMap()

	err = m.client.Create(ctx, &cfgMap)
	if err != nil {
		return fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	return fmt.Errorf("TODO: Do more stuff")
}
