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
	fmt.Println("Resolving config")
	cfgMap, err := m.cfg.PopulateFromConfigMap(ctx, m.client)
	if err != nil {
		return err
	}

	m.printConfig()
	fmt.Print(`
This will delete the old database instance. Rollback after this point is not possible.

Only proceed if you are sure that the migration was successful and that your application is working as expected.
`)

	err = confirmContinue()
	if err != nil {
		return err
	}

	jobName, err := m.doNaisJob(ctx, cfgMap, CommandCleanup)
	if err != nil {
		return err
	}

	label := m.kubectlLabelSelector(CommandCleanup)
	fmt.Printf(CleanupStartedMessage, label, m.cfg.Namespace)

	err = m.waitForJobCompletion(ctx, jobName, CommandCleanup)
	if err != nil {
		return err
	}

	err = m.deleteMigrationConfig(ctx)
	if err != nil {
		return err
	}

	fmt.Print(CleanupSuccessMessage)
	return nil
}
