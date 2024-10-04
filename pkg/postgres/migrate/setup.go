package migrate

import (
	"context"
	"fmt"
	"github.com/pterm/pterm"

	v1 "k8s.io/api/core/v1"
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

	pterm.Println("Resolving target instance config ...")
	err = m.cfg.Target.Resolve(ctx, m.client, m.cfg.AppName, m.cfg.Namespace)
	if err != nil {
		return err
	}

	pterm.Println("Resolving source instance config ...")
	err = m.cfg.Source.Resolve(ctx, m.client, m.cfg.AppName, m.cfg.Namespace)
	if err != nil {
		return err
	}

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

	m.printConfig()
	pterm.Warning.Println("This will create a new database instance and start replication of data from the source instance.")
	err = confirmContinue()
	if err != nil {
		return err
	}

	pterm.Println("Looking up GCP project ID ...")
	gcpProjectId, err := m.LookupGcpProjectId(ctx)
	if err != nil {
		return fmt.Errorf("failed to lookup GCP project ID: %w", err)
	}

	pterm.Println("Creating ConfigMap ...")
	cfgMap := m.cfg.CreateConfigMap()
	err = m.Create(ctx, cfgMap)
	if err != nil {
		return fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	pterm.Println("Creating RoleBinding ...")
	roleBinding := makeRoleBinding(m.cfg)
	err = createObject(ctx, m, cfgMap, roleBinding, CommandSetup)
	if err != nil {
		return err
	}

	jobName, err := m.doNaisJob(ctx, cfgMap, CommandSetup)
	if err != nil {
		return err
	}

	cloudConsoleUrl := fmt.Sprintf("https://console.cloud.google.com/dbmigration/migrations/locations/europe-north1/instances/%s-%s?project=%s", m.cfg.Source.InstanceName, m.cfg.Target.InstanceName, gcpProjectId)
	label := m.kubectlLabelSelector(CommandSetup)

	pterm.DefaultHeader.Println("Migration setup has been started successfully")
	pterm.Println()
	pterm.Println("To monitor the migration, run the following command:")
	cmdStyle.Printfln("\tkubectl logs -f -l %s -n %s", label, m.cfg.Namespace)
	pterm.Println()
	pterm.Println("The setup will take some time to complete, you can check completion status with the following command:")
	cmdStyle.Printfln("\tkubectl get job %s -n %s", jobName, m.cfg.Namespace)
	pterm.Println()
	pterm.Println("When setup is complete, a new instance has been created and replication of data has started.")
	pterm.Println("You can check the replication progress in the Google Cloud Console:")
	linkStyle.Printfln("\t%s", cloudConsoleUrl)
	pterm.Println()
	pterm.DefaultParagraph.Println("When the migration has status 'Running' and is in the 'CDC' or 'Ready to Promote' phase, you can proceed with the next step of the migration:")
	cmdStyle.Printfln("\tnais postgres migrate promote %s %s %s", m.cfg.AppName, m.cfg.Namespace, m.cfg.Target.InstanceName)
	pterm.Println()
	pterm.Info.Println("Be aware that during promotion (the next step), your instance will be unavailable for some time.")
	return nil
}
