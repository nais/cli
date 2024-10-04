package migrate

import (
	"context"
	"fmt"
	"github.com/pterm/pterm"
)

func (m *Migrator) Promote(ctx context.Context) error {
	pterm.Println("Resolving config ...")
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

	pterm.Println("Promotion has been started successfully.")
	pterm.Println()
	pterm.Println("To monitor the migration, run the following command:")
	cmdStyle.Printfln("\tkubectl logs -f -l %s -n %s", label, m.cfg.Namespace)
	pterm.Println()
	pterm.Println("The promote will take some time to complete, you can check completion status with the following command:")
	cmdStyle.Printfln("\tkubectl get job %s -n %s", jobName, m.cfg.Namespace)
	pterm.Println()
	pterm.Println("When promotion is complete, your application should be up and running with the new database instance.")
	pterm.Println()
	pterm.Info.Println(`At this point it is important to verify that your application works as expected, and that all data is present.
Once you are satisfied that everything works as expected, you must perform the final finalize step:`)
	cmdStyle.Printfln("\tnais postgres migrate finalize %s %s %s", m.cfg.AppName, m.cfg.Namespace, m.cfg.Target.InstanceName)
	pterm.Println()
	pterm.Info.Println("You must update your manifests to use the new database instance:")
	diskSizeLine := ""
	m.cfg.Target.DiskSize.Do(func(diskSize int) {
		diskSizeLine = fmt.Sprintf("diskSize: %d", diskSize)
	})
	yamlStyle.Printfln(`
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
	cmdStyle.Printfln("\tnais postgres migrate rollback %s %s %s", m.cfg.AppName, m.cfg.Namespace, m.cfg.Target.InstanceName)
	return nil
}
