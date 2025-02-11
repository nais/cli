package migrate

import (
	"context"
	"fmt"

	"github.com/nais/cli/pkg/postgres/migrate/ui"

	"github.com/pterm/pterm"
)

func (m *Migrator) Promote(ctx context.Context) error {
	cfgMap, err := m.cfg.PopulateFromConfigMap(ctx, m.client)
	if err != nil {
		return err
	}

	m.printConfig()
	pterm.Warning.Print(`Your application will not be able to reach the database during promotion.
The database will be unavailable for a short period of time while the promotion is in progress.
`)

	err = confirmContinue()
	if err != nil {
		return err
	}

	jobName, err := m.doNaisJob(ctx, cfgMap, CommandPromote)
	if err != nil {
		return err
	}

	label := m.kubectlLabelSelector(CommandPromote)

	if m.wait {
		printWaitingForJobHeader()
		err = m.waitForJobCompletion(ctx, jobName, CommandPromote)
		if err != nil {
			return err
		}

		pterm.Println()
		pterm.DefaultHeader.Println("Promotion completed successfully")
		pterm.Println()
		pterm.Println("Promotion is complete, your application should be up and running with the new database instance.")
	} else {
		pterm.Println()
		pterm.DefaultHeader.Println("Promotion has been started successfully")
		pterm.Println()
		pterm.Println("To monitor the migration, run the following command:")
		ui.CmdStyle.Printfln("\tkubectl logs -f -l %s", label)
		pterm.Println()
		pterm.Println("The promote will take some time to complete, you can check completion status with the following command:")
		ui.CmdStyle.Printfln("\tkubectl get job %s", jobName)
		pterm.Println()
		pterm.Println("When promotion is complete, your application should be up and running with the new database instance.")
	}

	pterm.Println()
	pterm.Info.Println(`At this point it is important to verify that your application works as expected, and that all data is present.
Once you are satisfied that everything works as expected, you must perform the final finalize step:`)
	ui.CmdStyle.Printfln("\tnais postgres migrate finalize %s %s", m.cfg.AppName, m.cfg.Target.InstanceName)
	pterm.Println()
	pterm.Info.Println("After completion of the finalize step, you must update your manifests to use the new database instance:")
	diskSizeLine := ""
	m.cfg.Target.DiskSize.Do(func(diskSize int) {
		diskSizeLine = fmt.Sprintf("diskSize: %d", diskSize)
	})
	ui.YamlStyle.Printfln(`
    ...
    spec:
      gcp:
        sqlInstances:
        - name: %s
          type: %s
          tier: %s
          %s
`, m.cfg.Target.InstanceName, m.cfg.Target.Type, m.cfg.Target.Tier, diskSizeLine)
	pterm.Println()
	pterm.Println("If things are not working as expected, and you need to rollback to the previous database instance, you can run:")
	ui.CmdStyle.Printfln("\tnais postgres migrate rollback %s %s", m.cfg.AppName, m.cfg.Target.InstanceName)
	return nil
}
