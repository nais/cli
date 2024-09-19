package migrate

import (
	"context"
	"fmt"
)

const CleanupSuccessMessage = `
Cleanup has been started successfully.

To monitor the cleanup, run the following command:
	kubectl logs -f -l %s -n %s

The cleanup will take some time to complete, you can check completion status with the following command:
	kubectl get job %s -n %s

When cleanup is complete, the old instance has been deleted and the migration is complete.
Congratulations, you're all done! 🎉
`

func (m *Migrator) Cleanup(ctx context.Context) error {
	jobName, err := m.doCommand(ctx, CommandCleanup)
	if err != nil {
		return err
	}

	label := m.kubectlLabelSelector(CommandCleanup)

	fmt.Printf(CleanupSuccessMessage, label, m.cfg.Namespace, jobName, m.cfg.Namespace)
	return nil
}
