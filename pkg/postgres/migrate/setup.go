package migrate

import (
	"context"
	"errors"
	"fmt"

	"github.com/nais/cli/pkg/option"
	"github.com/nais/cli/pkg/postgres/migrate/config"
	"github.com/nais/cli/pkg/postgres/migrate/ui"
	"github.com/nais/liberator/pkg/namegen"
	"github.com/pterm/pterm"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (m *Migrator) Setup(ctx context.Context) error {
	cfgMapList := &v1.ConfigMapList{}
	listOptions := []client.ListOption{
		client.InNamespace(m.cfg.Namespace),
		client.MatchingLabels{"migrator.nais.io/app-name": m.cfg.AppName},
	}
	err := m.client.List(ctx, cfgMapList, listOptions...)
	if err != nil {
		return err
	}

	if len(cfgMapList.Items) > 0 {
		return fmt.Errorf("migration config already exists for this application")
	}

	err = m.cfg.Source.Resolve(ctx, m.client, m.cfg.AppName, m.cfg.Namespace)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			pterm.Println()
			pterm.Error.Printfln("Application %s not found in namespace %s", m.cfg.AppName, m.cfg.Namespace)
			pterm.Println()
			pterm.Println("Set the correct namespace in your kubeconfig context, using this command:")
			ui.CmdStyle.Printfln("\tkubectl config set-context --current --namespace=<namespace>")
			pterm.Println()
			pterm.Println("Or specify the namespace with the --namespace flag")
			pterm.Println()
			return fmt.Errorf("app %s not found in namespace %s", m.cfg.AppName, m.cfg.Namespace)
		} else if errors.Is(err, config.ErrMissingSqlInstance) {
			pterm.Println()
			pterm.Error.Printfln("The Application %s does not have any SQL instances defined in the spec", m.cfg.AppName)
			pterm.Println()
		}
		return err
	}

	m.ConfigureTarget()

	err = m.cfg.Target.Resolve(ctx, m.client, m.cfg.AppName, m.cfg.Namespace)
	if err != nil {
		return err
	}

	m.clearDiskSizeIfDiskAutoresizeEnabled()

	err = m.validateInstanceNames()
	if err != nil {
		return err
	}

	m.printConfig()
	pterm.Warning.Println("Do not make structural database changes during migration!\nThis is not supported, and will cause problems!")
	err = confirmContinue()
	if err != nil {
		return err
	}

	gcpProjectId, err := m.LookupGcpProjectId(ctx)
	if err != nil {
		return fmt.Errorf("failed to lookup GCP project ID: %w", err)
	}

	cfgMap := m.cfg.CreateConfigMap()
	err = m.Create(ctx, cfgMap)
	if err != nil {
		return fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	roleBinding := makeRoleBinding(m.cfg)
	err = createObject(ctx, m, cfgMap, roleBinding, CommandSetup)
	if err != nil {
		return err
	}

	jobName, err := m.doNaisJob(ctx, cfgMap, CommandSetup)
	if err != nil {
		return err
	}

	// Make sure this logic is in sync with the corresponding logic in cloudsql-migrator...
	migrationJobName := fmt.Sprintf("%s-%s", m.cfg.Source.InstanceName, m.cfg.Target.InstanceName)
	maxlen := 60 // Google limit for migration job names
	if len(migrationJobName) > maxlen {
		var err error
		migrationJobName, err = namegen.ShortName(migrationJobName, maxlen)
		if err != nil {
			return fmt.Errorf("failed to shorten migration job name: %w", err)
		}
	}

	cloudConsoleUrl := fmt.Sprintf("https://console.cloud.google.com/dbmigration/migrations/locations/europe-north1/instances/%s?project=%s", migrationJobName, gcpProjectId)
	label := m.kubectlLabelSelector(CommandSetup)

	if m.wait {
		printWaitingForJobHeader()
		err = m.waitForJobCompletion(ctx, jobName, CommandSetup)
		if err != nil {
			return err
		}
		pterm.Println()
		pterm.DefaultHeader.Println("Migration setup completed successfully")
		pterm.Println()
		pterm.Println("Setup is now complete, a new instance has been created and replication of data has started.")
	} else {
		pterm.Println()
		pterm.DefaultHeader.Println("Migration setup has been started successfully")
		pterm.Println()
		pterm.Println("To monitor the migration, run the following command:")
		ui.CmdStyle.Printfln("\tkubectl logs -f -l %s", label)
		pterm.Println()
		pterm.Println("The setup will take some time to complete, you can check completion status with the following command:")
		ui.CmdStyle.Printfln("\tkubectl get job %s", jobName)
		pterm.Println()
		pterm.Println("When setup is complete, a new instance has been created and replication of data has started.")
	}

	pterm.Println("You can check the replication progress in the Google Cloud Console:")
	ui.LinkStyle.Printfln("\t%s", cloudConsoleUrl)
	pterm.Println()
	pterm.DefaultParagraph.Println("When the migration has status 'Running' and is in the 'CDC' or 'Ready to Promote' phase, you can proceed with the next step of the migration:")
	ui.CmdStyle.Printfln("\tnais postgres migrate promote %s %s", m.cfg.AppName, m.cfg.Target.InstanceName)
	pterm.Println()
	pterm.Info.Println("Be aware that during promotion (the next step), your instance will be unavailable for some time.")
	return nil
}

func (m *Migrator) validateInstanceNames() error {
	sourceInstanceName := m.cfg.Source.InstanceName.String()
	if sourceInstanceName == "" {
		return fmt.Errorf("source instance name is empty")
	}

	targetInstanceName := m.cfg.Target.InstanceName.String()
	if targetInstanceName == "" {
		return fmt.Errorf("target instance name is required")
	}

	if sourceInstanceName == targetInstanceName {
		return fmt.Errorf("source and target instance names cannot be the same")
	}
	return nil
}

func (m *Migrator) clearDiskSizeIfDiskAutoresizeEnabled() {
	m.cfg.Target.DiskAutoresize.Do(func(v bool) {
		if v {
			m.cfg.Target.DiskSize = option.None[int]()
		}
	})
}

func (m *Migrator) ConfigureTarget() {
	m.cfg.Target.Tier = m.cfg.Target.Tier.OrMaybe(ui.AskForTier(m.cfg.Source.Tier.String()))
	m.cfg.Target.Type = m.cfg.Target.Type.OrMaybe(ui.AskForType(m.cfg.Source.Type.String()))
	m.cfg.Target.DiskAutoresize = m.cfg.Target.DiskAutoresize.OrMaybe(ui.AskForDiskAutoresize(m.cfg.Source.DiskAutoresize))
	m.cfg.Target.DiskAutoresize.Do(func(v bool) {
		if !v {
			m.cfg.Target.DiskSize = m.cfg.Target.DiskSize.OrMaybe(ui.AskForDiskSize(m.cfg.Source.DiskSize))
		}
	})
}
