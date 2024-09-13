package migrate

import (
	"context"
	"fmt"
)

func (m *Migrator) Promote(ctx context.Context) error {
	fmt.Println("Resolving config")
	cfgMap, err := m.cfg.PopulateFromConfigMap(ctx, m.client)
	if err != nil {
		return err
	}

	fmt.Println("Creating NaisJob")
	imageTag, err := getLatestImageTag()
	if err != nil {
		return fmt.Errorf("failed to get latest image tag for cloudsql-migrator: %w", err)
	}
	job := makeNaisjob(m.cfg, imageTag, CommandPromote)
	err = createObject(ctx, m, cfgMap, job)
	if err != nil {
		return err
	}

	return nil
}
