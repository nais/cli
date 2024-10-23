package migrate

import (
	"context"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

func (m *Migrator) Finalize(ctx context.Context) error {
	cfgMap, err := m.cfg.PopulateFromConfigMap(ctx, m.client)
	if err != nil {
		return err
	}

	m.printConfig()
	pterm.Warning.Print(`This will delete the old database instance. Rollback after this point is not possible.
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

	err = m.waitForJobCompletion(ctx, jobName, CommandFinalize)
	if err != nil {
		return err
	}

	err = m.deleteMigrationConfig(ctx, cfgMap)
	if err != nil {
		return err
	}

	pterm.Println()
	pterm.DefaultHeader.Println("Finalize has completed successfully")
	pterm.Println()
	pterm.Println("The old instance has been deleted and the migration is complete.")
	pterm.Println()
	pterm.DefaultBigText.WithLetters(putils.LettersFromString("Congrats!")).Render()
	pterm.Println("You are all done! 🎉")

	return nil
}
