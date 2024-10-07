package migrate

import (
	"context"
	"fmt"
	"github.com/nais/cli/pkg/option"
	"github.com/pterm/pterm"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
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
		if errors.IsNotFound(err) {
			pterm.Println()
			pterm.Error.Printfln("Application %s not found in namespace %s", m.cfg.AppName, m.cfg.Namespace)
			pterm.Println()
			pterm.Println("Set the correct namespace in your kubeconfig context, using this command:")
			cmdStyle.Printfln("\tkubectl config set-context --current --namespace=<namespace>")
			pterm.Println()
			pterm.Println("Or specify the namespace with the --namespace flag")
			pterm.Println()
			return fmt.Errorf("app %s not found in namespace %s", m.cfg.AppName, m.cfg.Namespace)
		}
		return err
	}

	m.cfg.Target.Tier = m.cfg.Target.Tier.OrMaybe(askForTier(m.cfg.Source.Tier.String()))
	m.cfg.Target.Type = m.cfg.Target.Type.OrMaybe(askForType(m.cfg.Source.Type.String()))
	m.cfg.Target.DiskSize = m.cfg.Target.DiskSize.OrMaybe(askForDiskSize)

	err = m.cfg.Target.Resolve(ctx, m.client, m.cfg.AppName, m.cfg.Namespace)
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

	cloudConsoleUrl := fmt.Sprintf("https://console.cloud.google.com/dbmigration/migrations/locations/europe-north1/instances/%s-%s?project=%s", m.cfg.Source.InstanceName, m.cfg.Target.InstanceName, gcpProjectId)
	label := m.kubectlLabelSelector(CommandSetup)

	pterm.DefaultHeader.Println("Migration setup has been started successfully")
	pterm.Println()
	pterm.Println("To monitor the migration, run the following command:")
	cmdStyle.Printfln("\tkubectl logs -f -l %s", label)
	pterm.Println()
	pterm.Println("The setup will take some time to complete, you can check completion status with the following command:")
	cmdStyle.Printfln("\tkubectl get job %s", jobName)
	pterm.Println()
	pterm.Println("When setup is complete, a new instance has been created and replication of data has started.")
	pterm.Println("You can check the replication progress in the Google Cloud Console:")
	linkStyle.Printfln("\t%s", cloudConsoleUrl)
	pterm.Println()
	pterm.DefaultParagraph.Println("When the migration has status 'Running' and is in the 'CDC' or 'Ready to Promote' phase, you can proceed with the next step of the migration:")
	cmdStyle.Printfln("\tnais postgres migrate promote %s %s", m.cfg.AppName, m.cfg.Target.InstanceName)
	pterm.Println()
	pterm.Info.Println("Be aware that during promotion (the next step), your instance will be unavailable for some time.")
	return nil
}

const otherOption = "Other"

var tierOptions = []string{
	"db-custom-1-3840",
	"db-custom-2-5120",
	"db-custom-2-7680",
	"db-custom-4-15360",
	otherOption,
}

func askForTier(sourceTier string) func() option.Option[string] {
	return func() option.Option[string] {
		options := []string{"Same as source"}
		for _, tier := range tierOptions {
			if tier != sourceTier {
				options = append(options, tier)
			}
		}
		pterm.Println()
		tier, err := pterm.DefaultInteractiveSelect.
			WithOptions(options).
			WithMaxHeight(len(options)).
			Show("Select a tier for the target instance")
		if err != nil {
			log.Fatalf("Error while creating text UI: %v", err)
			return option.None[string]()
		}
		if tier == otherOption {
			pterm.Println("Check the documentation for possible options:")
			linkStyle.Printfln("\thttps://doc.nais.io/persistence/postgres/reference/#server-size")
			tier, err = pterm.DefaultInteractiveTextInput.Show("Enter the tier for the target instance")
			if err != nil {
				log.Fatalf("Error while creating text UI: %v", err)
				return option.None[string]()
			}
		}
		return option.Some(tier)
	}
}

var typeToVersion = map[string]int{
	"POSTGRES_11": 11,
	"POSTGRES_12": 12,
	"POSTGRES_13": 13,
	"POSTGRES_14": 14,
	"POSTGRES_15": 15,
	"POSTGRES_16": 16,
}

func askForType(sourceType string) func() option.Option[string] {
	sourceVersion := typeToVersion[sourceType]
	return func() option.Option[string] {
		options := []string{"Same as source"}
		for k, v := range typeToVersion {
			if v > sourceVersion {
				options = append(options, k)
			}
		}
		if len(options) == 1 {
			return option.None[string]()
		}
		pterm.Println()
		instanceType, err := pterm.DefaultInteractiveSelect.
			WithOptions(options).
			WithMaxHeight(len(options)).
			Show("Select a type for the target instance")
		if err != nil {
			log.Fatalf("Error while creating text UI: %v", err)
			return option.None[string]()
		}
		if instanceType == "Same as source" {
			return option.None[string]()
		}
		return option.Some(instanceType)
	}
}

func askForDiskSize() option.Option[int] {
	pterm.Println()
	pterm.Println("Disk size is in GB, and must be greater than or equal to 10.")
	diskSize, err := pterm.DefaultInteractiveTextInput.Show("Enter the disk size for the target instance. Leave empty to use same as source")
	if err != nil {
		log.Fatalf("Error while creating text UI: %v", err)
		return option.None[int]()
	}
	if diskSize == "" {
		return option.None[int]()
	}
	size, err := strconv.Atoi(diskSize)
	if err != nil {
		pterm.Error.Println("Disk size must be a number")
		return askForDiskSize()
	}
	if size < 10 {
		pterm.Error.Println("Disk size must be greater than or equal to 10")
		return askForDiskSize()
	}
	return option.Some(size)
}
