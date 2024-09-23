package migrate

import (
	"context"
	"fmt"
)

const CleanupStartedMessage = `
Cleanup has been started successfully.

To monitor the cleanup, run the following command in a separate terminal:
	kubectl logs -f -l %s -n %s

Pausing to wait for cleanup job to complete in order to do final cleanup actions ...
`

const CleanupSuccessMessage = `
Cleanup has completed successfully.

The old instance has been deleted and the migration is complete.

Congratulations, you're all done! 🎉
`

func (m *Migrator) Cleanup(ctx context.Context) error {
	jobName, err := m.doCommand(ctx, CommandCleanup)
	if err != nil {
		return err
	}

	label := m.kubectlLabelSelector(CommandCleanup)
	fmt.Printf(CleanupStartedMessage, label, m.cfg.Namespace)

	err = m.waitForJobCompletion(ctx, jobName, CommandCleanup)
	if err != nil {
		return err
	}

	fmt.Print(CleanupSuccessMessage)
	return nil
}
