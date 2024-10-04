package migrate

import (
	"context"
	"fmt"
	"github.com/pterm/pterm"
)

const RollbackStartedMessage = `
Rollback has been started successfully.

To monitor the rollback, run the following command in a separate terminal:
	kubectl logs -f -l %s -n %s

Pausing to wait for rollback job to complete in order to do final finalize actions ...
`

const RollbackSuccessMessage = `
Rollback has completed successfully.

Your application should be up and running with the original database instance.
The new instance has been deleted and the migration is stopped.

You are now free to start another attempt if you wish.
`

func (m *Migrator) Rollback(ctx context.Context) error {
	pterm.Println("Resolving config ...")
	cfgMap, err := m.cfg.PopulateFromConfigMap(ctx, m.client)
	if err != nil {
		return err
	}

	m.printConfig()
	pterm.Warning.Println("This will roll back the migration, and restore the application to use the original instance.")

	err = confirmContinue()
	if err != nil {
		return err
	}

	jobName, err := m.doNaisJob(ctx, cfgMap, CommandRollback)
	if err != nil {
		return err
	}

	label := m.kubectlLabelSelector(CommandRollback)
	fmt.Printf(RollbackStartedMessage, label, m.cfg.Namespace)

	err = m.waitForJobCompletion(ctx, jobName, CommandRollback)
	if err != nil {
		return err
	}

	err = m.deleteMigrationConfig(ctx)
	if err != nil {
		return err
	}

	fmt.Print(RollbackSuccessMessage)
	return nil
}
