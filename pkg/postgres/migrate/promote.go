package migrate

import (
	"context"
	"fmt"
)

const PromoteSuccessMessage = `
Promotion has been started successfully.

To monitor the migration, run the following command:
	kubectl logs -f -l %s -n %s

The promote will take some time to complete, you can check completion status with the following command:
	kubectl get job %s -n %s

When promotion is complete, your application should be up and running with the new database instance.
At this point it is important to verify that your application works as expected, and that all data is present.

Once you are satisfied that everything works as expected, you must perform the final cleanup step:
	nais postgres migrate cleanup %s %s %s

If things are not working as expected, and you need to rollback to the previous database instance, you can run:
	nais postgres migrate rollback %s %s %s
`

func (m *Migrator) Promote(ctx context.Context) error {
	jobName, err := m.doCommand(ctx, CommandPromote)
	if err != nil {
		return err
	}

	label := m.kubectlLabelSelector(CommandPromote)
	fmt.Printf(PromoteSuccessMessage, label, m.cfg.Namespace, jobName, m.cfg.Namespace, m.cfg.AppName, m.cfg.Namespace, m.cfg.Target.InstanceName, m.cfg.AppName, m.cfg.Namespace, m.cfg.Target.InstanceName)
	return nil
}
