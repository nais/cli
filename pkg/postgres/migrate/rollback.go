package migrate

import (
	"context"
	"github.com/pterm/pterm"
)

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
	pterm.DefaultHeader.Println("Rollback has been started successfully")
	pterm.Println()
	pterm.Println("To monitor the rollback, run the following command in a separate terminal:")
	cmdStyle.Printfln("\tkubectl logs -f -l %s", label)
	pterm.Println()

	err = m.waitForJobCompletion(ctx, jobName, CommandRollback)
	if err != nil {
		return err
	}

	err = m.deleteMigrationConfig(ctx)
	if err != nil {
		return err
	}

	pterm.DefaultHeader.Println("Rollback has completed successfully")
	pterm.Println()
	pterm.Println("Your application should be up and running with the original database instance.")
	pterm.Println("The new instance has been deleted and the migration is stopped.")
	pterm.Println()
	pterm.Println("You are now free to start another attempt if you wish.")

	return nil
}
