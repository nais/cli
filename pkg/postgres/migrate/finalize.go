package migrate

import (
	"context"
	"fmt"
)

const FinalizeStartedMessage = `
Finalize has been started successfully.

To monitor the finalize, run the following command in a separate terminal:
	kubectl logs -f -l %s -n %s

Pausing to wait for finalize job to complete in order to do final finalize actions ...
`

const FinalizeSuccessMessage = `
Finalize has completed successfully.

The old instance has been deleted and the migration is complete.

Congratulations, you're all done! 🎉
`

func (m *Migrator) Finalize(ctx context.Context) error {
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

	jobName, err := m.doNaisJob(ctx, cfgMap, CommandFinalize)
	if err != nil {
		return err
	}

	label := m.kubectlLabelSelector(CommandFinalize)
	fmt.Printf(FinalizeStartedMessage, label, m.cfg.Namespace)

	err = m.waitForJobCompletion(ctx, jobName, CommandFinalize)
	if err != nil {
		return err
	}

	err = m.deleteMigrationConfig(ctx)
	if err != nil {
		return err
	}

	fmt.Print(FinalizeSuccessMessage)
	return nil
}
