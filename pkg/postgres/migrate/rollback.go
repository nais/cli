package migrate

import (
	"context"
	"fmt"
)

const RollbackSuccessMessage = `
Rollback has been started successfully.

To monitor the rollback, run the following command:
	kubectl logs -f -l %s -n %s

The rollback will take some time to complete, you can check completion status with the following command:
	kubectl get job %s -n %s

When rollback is complete, your application should be up and running with the original database instance.
`

func (m *Migrator) Rollback(ctx context.Context) error {
	jobName, err := m.doCommand(ctx, CommandRollback)
	if err != nil {
		return err
	}

	label := m.kubectlLabelSelector(CommandRollback)
	fmt.Printf(RollbackSuccessMessage, label, m.cfg.Namespace, jobName, m.cfg.Namespace)
	return nil
}
