package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pterm/pterm"
	"net/http"
	"reflect"
	"time"

	"github.com/nais/cli/pkg/postgres/migrate/config"
	"github.com/sethvargo/go-retry"

	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Command string

func (c Command) JobName(cfg config.Config) string {
	return fmt.Sprintf("%s-%s", cfg.MigrationName(), string(c))
}

const (
	CommandFinalize Command = "finalize"
	CommandPromote  Command = "promote"
	CommandRollback Command = "rollback"
	CommandSetup    Command = "setup"
)

const MigratorImage = "europe-north1-docker.pkg.dev/nais-io/nais/images/cloudsql-migrator"

var cmdStyle = pterm.NewStyle(pterm.FgLightMagenta)
var linkStyle = pterm.NewStyle(pterm.FgLightBlue, pterm.Underscore)
var yamlStyle = pterm.NewStyle(pterm.FgLightYellow)

type Migrator struct {
	client ctrl.Client
	cfg    config.Config
	dryRun bool
	wait   bool
}

func NewMigrator(client ctrl.Client, cfg config.Config, dryRun bool, noWait bool) *Migrator {
	return &Migrator{
		client: client,
		cfg:    cfg,
		dryRun: dryRun,
		wait:   !noWait,
	}
}

func (m *Migrator) Create(ctx context.Context, obj ctrl.Object) error {
	if m.dryRun {
		v := reflect.Indirect(reflect.ValueOf(obj))
		pterm.Printf("Dry run: Skipping creation of %s: %s\n", v.Type().Name(), obj.GetName())
		return nil
	}
	return m.client.Create(ctx, obj)
}

func (m *Migrator) Delete(ctx context.Context, obj ctrl.Object) error {
	if m.dryRun {
		v := reflect.Indirect(reflect.ValueOf(obj))
		pterm.Printf("Dry run: Skipping deletion of %s: %s\n", v.Type().Name(), obj.GetName())
		return nil
	}
	return m.client.Delete(ctx, obj)
}

func (m *Migrator) doNaisJob(ctx context.Context, cfgMap *corev1.ConfigMap, command Command) (string, error) {
	imageTag, err := getLatestImageTag()
	if err != nil {
		return "", fmt.Errorf("failed to get latest image tag for cloudsql-migrator: %w", err)
	}

	job := makeNaisjob(m.cfg, imageTag, command)
	err = createObject(ctx, m, cfgMap, job, command)
	if err != nil {
		return "", err
	}

	return job.Name, nil
}

func (m *Migrator) kubectlLabelSelector(command Command) string {
	return fmt.Sprintf("migrator.nais.io/migration-name=%s,migrator.nais.io/command=%s", m.cfg.MigrationName(), command)
}

func (m *Migrator) deleteMigrationConfig(ctx context.Context) error {
	err := ctrl.IgnoreNotFound(m.Delete(ctx, m.cfg.GetConfigMap()))
	if err != nil {
		return fmt.Errorf("failed to delete ConfigMap: %w", err)
	}

	return nil
}

func (m *Migrator) LookupGcpProjectId(ctx context.Context) (string, error) {
	ns := &corev1.Namespace{}
	err := m.client.Get(ctx, ctrl.ObjectKey{Name: m.cfg.Namespace}, ns)
	if err != nil {
		return "", fmt.Errorf("failed to get namespace: %w", err)
	}
	if gcpProjectId, ok := ns.Annotations["cnrm.cloud.google.com/project-id"]; ok {
		return gcpProjectId, nil
	}
	return "", fmt.Errorf("namespace %s does not have a GCP project ID annotation", m.cfg.Namespace)
}

func (m *Migrator) waitForJobCompletion(ctx context.Context, jobName string, command Command) error {
	multi := pterm.DefaultMultiPrinter
	logOutput := pterm.DefaultLogger.WithMaxWidth(120).WithWriter(multi.NewWriter())
	logOutput.Info(fmt.Sprintf("Pausing to wait for %s job to complete in order to do final cleanup actions ...", command))
	spinner, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Waiting for job to complete ...")
	multi.Start()
	defer multi.Stop()

	if m.dryRun {
		logOutput.Info(fmt.Sprintf("Dry run: Artificial waiting for job %s/%s to complete, 5 seconds\n", m.cfg.Namespace, jobName))
		time.Sleep(5 * time.Second)
		spinner.Success("Job completed")
		return nil
	}
	listOptions := []ctrl.ListOption{
		ctrl.InNamespace(m.cfg.Namespace),
		ctrl.MatchingLabels{
			"migrator.nais.io/migration-name": m.cfg.MigrationName(),
			"migrator.nais.io/command":        string(command),
		},
	}

	b := retry.NewConstant(10 * time.Second)
	b = retry.WithMaxDuration(15*time.Minute, b)
	err := retry.Do(ctx, b, func(ctx context.Context) error {
		jobs := &batchv1.JobList{}
		err := m.client.List(ctx, jobs, listOptions...)
		if err != nil {
			logOutput.Warn(fmt.Sprintf("Error getting jobs %s/%s, retrying: %v\n", m.cfg.Namespace, jobName, err))
			return retry.RetryableError(err)
		}
		if len(jobs.Items) < 1 {
			return retry.RetryableError(fmt.Errorf("no jobs found"))
		}
		if len(jobs.Items) > 1 {
			logOutput.Error(fmt.Sprintf("Multiple jobs found %s/%s! This should not happen, contact the nais team!\n", m.cfg.Namespace, jobName))
			return fmt.Errorf("multiple jobs found")
		}
		for _, job := range jobs.Items {
			if job.Status.Succeeded == 1 {
				return nil
			}
		}
		return retry.RetryableError(fmt.Errorf("job %s/%s has not completed yet", m.cfg.Namespace, jobName))
	})
	if err != nil {
		spinner.Fail()
		return err
	}
	spinner.Success("Job completed")
	return nil
}

func (m *Migrator) printConfig() {
	pterm.DefaultSection.Println("Migration configuration")
	pterm.Printfln("Application: %s", m.cfg.AppName)
	pterm.Printfln("Namespace: %s", m.cfg.Namespace)
	pterm.DefaultSection.Println("Instance configuration")
	sourceDiskSize := "<nais default>"
	m.cfg.Source.DiskSize.Do(func(diskSize int) {
		sourceDiskSize = fmt.Sprintf("%d GB", diskSize)
	})
	targetDiskSize := "<nais default>"
	m.cfg.Target.DiskSize.Do(func(diskSize int) {
		targetDiskSize = fmt.Sprintf("%d GB", diskSize)
	})

	tableHeaderStyle := pterm.ThemeDefault.TableHeaderStyle
	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"", "Name", "Tier", "Disk size", "Type"},
		{tableHeaderStyle.Sprint("Source"), m.cfg.Source.InstanceName.String(), m.cfg.Source.Tier.String(), sourceDiskSize, m.cfg.Source.Type.String()},
		{tableHeaderStyle.Sprint("Target"), m.cfg.Target.InstanceName.String(), m.cfg.Target.Tier.String(), targetDiskSize, m.cfg.Target.Type.String()},
	}).Render()
}

func createObject[T interface {
	ctrl.Object
	*P
}, P any](ctx context.Context, m *Migrator, owner metav1.Object, obj T, Command Command) error {
	err := controllerutil.SetOwnerReference(owner, obj, m.client.Scheme())
	if err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["migrator.nais.io/migration-name"] = m.cfg.MigrationName()
	labels["migrator.nais.io/app-name"] = m.cfg.AppName
	labels["migrator.nais.io/target-instance-name"] = m.cfg.Target.InstanceName.String()
	labels["migrator.nais.io/command"] = string(Command)
	obj.SetLabels(labels)

	err = m.Create(ctx, obj)
	if err != nil {
		return fmt.Errorf("failed to create Object: %w", err)
	}
	return nil
}

func makeRoleBinding(cfg config.Config) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.MigrationName(),
			Namespace: cfg.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: CommandSetup.JobName(cfg),
			},
			{
				Kind: "ServiceAccount",
				Name: CommandPromote.JobName(cfg),
			},
			{
				Kind: "ServiceAccount",
				Name: CommandFinalize.JobName(cfg),
			},
			{
				Kind: "ServiceAccount",
				Name: CommandRollback.JobName(cfg),
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "nais:developer",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func makeNaisjob(cfg config.Config, imageTag string, command Command) *nais_io_v1.Naisjob {
	return &nais_io_v1.Naisjob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      command.JobName(cfg),
			Namespace: cfg.Namespace,
			Labels: map[string]string{
				"apiserver-access": "enabled",
			},
		},
		Spec: nais_io_v1.NaisjobSpec{
			Command: []string{"/" + string(command)},
			Env: nais_io_v1.EnvVars{
				{
					Name:  "LOG_FORMAT",
					Value: "JSON",
				},
			},
			EnvFrom: []nais_io_v1.EnvFrom{{
				ConfigMap: cfg.MigrationName(),
			}},
			GCP: &nais_io_v1.GCP{
				Permissions: []nais_io_v1.CloudIAMPermission{
					{
						Role: "roles/cloudsql.admin",
						Resource: nais_io_v1.CloudIAMResource{
							APIVersion: "resourcemanager.cnrm.cloud.google.com/v1beta1",
							Kind:       "Project",
						},
					}, {
						Role: "roles/datamigration.admin",
						Resource: nais_io_v1.CloudIAMResource{
							APIVersion: "resourcemanager.cnrm.cloud.google.com/v1beta1",
							Kind:       "Project",
						},
					}, {
						Role: "roles/monitoring.viewer",
						Resource: nais_io_v1.CloudIAMResource{
							APIVersion: "resourcemanager.cnrm.cloud.google.com/v1beta1",
							Kind:       "Project",
						},
					},
				},
			},
			Image: fmt.Sprintf("%s:%s", MigratorImage, imageTag),
		},
	}
}

func getLatestImageTag() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/nais/cloudsql-migrator/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get latest release: %s", resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)
	v := map[string]interface{}{}
	err = decoder.Decode(&v)
	if err != nil {
		return "", err
	}

	return v["tag_name"].(string), nil
}

func confirmContinue() error {
	pterm.Println()
	result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to continue?")
	pterm.Println()

	if !result {
		return fmt.Errorf("cancelled by user")
	}

	return nil
}
