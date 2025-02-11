package migrate

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/nais/cli/pkg/postgres/migrate/config"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"github.com/nais/liberator/pkg/namegen"
	"github.com/pterm/pterm"
	"github.com/sethvargo/go-retry"
	"golang.org/x/sync/errgroup"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Command string

func (c Command) JobName(cfg config.Config) string {
	base := cfg.MigrationName()
	suffix := string(c)
	name := fmt.Sprintf("%s-%s", base, suffix)
	maxlen := validation.DNS1123LabelMaxLength

	if len(name) > maxlen {
		truncated, err := namegen.SuffixedShortName(base, suffix, maxlen)
		if err != nil {
			panic(fmt.Sprintf("BUG: generating job name: %v", err.Error()))
		}
		return truncated
	}

	return name
}

const (
	CommandFinalize Command = "finalize"
	CommandPromote  Command = "promote"
	CommandRollback Command = "rollback"
	CommandSetup    Command = "setup"
)

const MigratorImage = "europe-north1-docker.pkg.dev/nais-io/nais/images/cloudsql-migrator"

type logEntry struct {
	Msg                 string `json:"msg"`
	Level               string `json:"level"`
	MigrationStep       int    `json:"migrationStep"`
	MigrationStepsTotal int    `json:"migrationStepsTotal"`
	extra               map[string]any
}

var irrelevantExtraLogEntryKeys = []string{
	"msg",
	"time",
	"level",
	"source",
	"migrationApp",
	"migrationTarget",
	"migrationPhase",
	"migrationStep",
	"migrationStepsTotal",
	"config",
}

type Migrator struct {
	client    ctrl.Client
	clientset kubernetes.Interface
	cfg       config.Config
	dryRun    bool
	wait      bool
}

func NewMigrator(client ctrl.Client, clientset kubernetes.Interface, cfg config.Config, dryRun bool, noWait bool) *Migrator {
	return &Migrator{
		client:    client,
		clientset: clientset,
		cfg:       cfg,
		dryRun:    dryRun,
		wait:      !noWait,
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

func (m *Migrator) deleteMigrationConfig(ctx context.Context, cfgMap *corev1.ConfigMap) error {
	err := ctrl.IgnoreNotFound(m.Delete(ctx, cfgMap))
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

func (m *Migrator) getJobLogs(ctx context.Context, command Command, jobName string, logChannel chan<- string) error {
	defer close(logChannel)

	if m.dryRun {
		send := func(entry logEntry) {
			entry.Msg = fmt.Sprintf("Dry run: %s", entry.Msg)
			b, _ := json.Marshal(entry)
			logChannel <- string(b)
			time.Sleep(500 * time.Millisecond)
		}
		send(logEntry{Msg: fmt.Sprintf("Starting %s", command), Level: "info", MigrationStep: 1, MigrationStepsTotal: 3})
		send(logEntry{Msg: "Running", Level: "info", MigrationStep: 2})
		send(logEntry{Msg: "Simulating log output", Level: "info"})
		send(logEntry{Msg: "Simulating more log output", Level: "warn"})
		send(logEntry{Msg: "Simulating even more log output", Level: "error"})
		send(logEntry{Msg: "Job completed", Level: "info"})
		send(logEntry{Msg: "Finished", Level: "info", MigrationStep: 3})
		return nil
	}

	seenPods := make(map[string]bool)

	for ctx.Err() == nil {
		pod, err := m.findLatestPod(ctx, command)
		if err != nil {
			return fmt.Errorf("error finding pod: %w", err)
		}

		switch {
		case pod == nil:
			// No pod found; wait for it to be created.
			time.Sleep(1 * time.Second)
			continue
		case pod.Status.Phase == corev1.PodSucceeded:
			// Pod (and thus Job) has completed successfully.
			logChannel <- `{"msg": ">>> Pod succeeded", "level": "info", "pod": "` + pod.Name + `"}`
			return nil
		case pod.Status.Phase != corev1.PodRunning:
			// Pod is not running yet; wait for it to start.
			logChannel <- `{"msg": ">>> Pod not running yet, waiting...", "level": "info", "pod": "` + pod.Name + `", "phase": "` + string(pod.Status.Phase) + `"}`
			time.Sleep(1 * time.Second)
			continue
		case seenPods[pod.Name]:
			// We've already printed logs for this pod; wait for a new pod to be created.
			time.Sleep(1 * time.Second)
			continue
		}

		logs, err := m.clientset.CoreV1().Pods(m.cfg.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			Container: jobName,
			Follow:    true,
		}).Stream(ctx)
		if err != nil {
			return fmt.Errorf("error getting job logs: %w", err)
		}

		logChannel <- `{"msg": ">>> Log stream started", "level": "info", "pod": "` + pod.Name + `"}`
		scanner := bufio.NewScanner(logs)
		for scanner.Scan() {
			logChannel <- scanner.Text()
		}
		logs.Close()
		logChannel <- `{"msg": ">>> Log stream ended", "level": "info", "pod": "` + pod.Name + `"}`

		// The stream ended, which likely means the pod either exited (whether successful or not) or was deleted.
		// Mark the pod as seen to avoid printing its logs again.
		seenPods[pod.Name] = true

		err = scanner.Err()
		if err != nil {
			return fmt.Errorf("error reading job logs: %w", err)
		}
	}
	return nil
}

// findLatestPod returns the latest pod for the given command. If no pods are found, nil is returned.
func (m *Migrator) findLatestPod(ctx context.Context, command Command) (*corev1.Pod, error) {
	pods, err := m.clientset.CoreV1().Pods(m.cfg.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: m.kubectlLabelSelector(command),
	})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	var latest *corev1.Pod
	latestTime := metav1.Time{}

	for _, pod := range pods.Items {
		if pod.GetCreationTimestamp().After(latestTime.Time) {
			latest = &pod
			latestTime = pod.GetCreationTimestamp()
		}
	}

	return latest, nil
}

func (m *Migrator) waitForJobCompletion(ctx context.Context, jobName string, command Command) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logChannel := make(chan string)

	// ctx is now canceled if any goroutine within the errgroup returns an error, or all of them complete successfully.
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return m.getJobLogs(ctx, command, jobName, logChannel)
	})

	startingMessage, err := m.waitForStartingMessage(ctx, logChannel)
	if err != nil {
		return err
	}

	logOutput := pterm.DefaultLogger.WithMaxWidth(120)
	logOutput.Info(startingMessage.Msg)

	progress, _ := pterm.DefaultProgressbar.WithTotal(startingMessage.MigrationStepsTotal).WithMaxWidth(120).Start()
	defer progress.Stop()

	// this runs outside the errgroup as it does not return an error
	go renderJobLogs(ctx, logChannel, logOutput, progress)

	if m.dryRun {
		logOutput.Info(fmt.Sprintf("Dry run: Artificial waiting for job %s/%s to complete, 5 seconds\n", m.cfg.Namespace, jobName))
		time.Sleep(5 * time.Second)
		return nil
	}

	eg.Go(func() error {
		return m.pollJobCompletion(ctx, jobName, command)
	})

	err = eg.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			err = context.Cause(ctx)
		}
		logOutput.Error(err.Error())
		return fmt.Errorf("error waiting for job completion: %w", err)
	}

	return nil
}

func (m *Migrator) waitForStartingMessage(ctx context.Context, logChannel <-chan string) (*logEntry, error) {
	spinner, _ := pterm.DefaultSpinner.Start("Waiting for job to start ...")
	defer spinner.Stop()

	for {
		select {
		case <-ctx.Done():
			spinner.Fail()
			err := context.Cause(ctx)
			pterm.Error.Println(err)
			return nil, err
		case line := <-logChannel:
			l, err := parseLogLine(line)
			if err != nil {
				spinner.Fail()
				pterm.Error.Println(err)
				return nil, err
			}

			if l.MigrationStepsTotal > 0 {
				return &l, nil
			}
		}
	}
}

func (m *Migrator) pollJobCompletion(ctx context.Context, jobName string, command Command) error {
	listOptions := []ctrl.ListOption{
		ctrl.InNamespace(m.cfg.Namespace),
		ctrl.MatchingLabels{
			"migrator.nais.io/migration-name": m.cfg.MigrationName(),
			"migrator.nais.io/command":        string(command),
		},
	}

	b := retry.NewConstant(10 * time.Second)
	return retry.Do(ctx, b, func(ctx context.Context) error {
		jobs := &batchv1.JobList{}
		err := m.client.List(ctx, jobs, listOptions...)
		if err != nil {
			return retry.RetryableError(err)
		}
		if len(jobs.Items) < 1 {
			return retry.RetryableError(fmt.Errorf("no jobs found"))
		}
		if len(jobs.Items) > 1 {
			return fmt.Errorf("multiple jobs found %s/%s, contact nais team", m.cfg.Namespace, jobName)
		}
		for _, job := range jobs.Items {
			if job.Status.Succeeded == 1 {
				return nil
			}
		}
		return retry.RetryableError(fmt.Errorf("job %s/%s has not completed yet", m.cfg.Namespace, jobName))
	})
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
	sourceAutoresize := "<nais default>"
	m.cfg.Source.DiskAutoresize.Do(func(autoresize bool) {
		if autoresize {
			sourceAutoresize = "enabled"
		} else {
			sourceAutoresize = "disabled"
		}
	})
	targetAutoresize := "<nais default>"
	m.cfg.Target.DiskAutoresize.Do(func(autoresize bool) {
		if autoresize {
			targetAutoresize = "enabled"
		} else {
			targetAutoresize = "disabled"
		}
	})

	tableHeaderStyle := pterm.ThemeDefault.TableHeaderStyle
	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"", "Name", "Tier", "Disk autoresize", "Disk size", "Type"},
		{tableHeaderStyle.Sprint("Source"), m.cfg.Source.InstanceName.String(), m.cfg.Source.Tier.String(), sourceAutoresize, sourceDiskSize, m.cfg.Source.Type.String()},
		{tableHeaderStyle.Sprint("Target"), m.cfg.Target.InstanceName.String(), m.cfg.Target.Tier.String(), targetAutoresize, targetDiskSize, m.cfg.Target.Type.String()},
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

func printWaitingForJobHeader() {
	pterm.Println("Several of the operations done by the migrator are eventually consistent, and may fail a few times before succeeding.")
	pterm.Println("This leads to some log messages about errors or failures, but the operations will typically be retried and eventually succeed.")
	pterm.Println("If there is an unrecoverable error, the migrator will exit with an error message.")
}

func parseLogLine(line string) (logEntry, error) {
	var le logEntry
	err := json.Unmarshal([]byte(line), &le)
	if err != nil {
		return logEntry{}, err
	}

	// pick up additional log fields that are not part of the logEntry struct
	extra := make(map[string]any)
	// this error should be caught above in previous Unmarshal
	_ = json.Unmarshal([]byte(line), &extra)

	for _, key := range irrelevantExtraLogEntryKeys {
		delete(extra, key)
	}

	le.extra = extra
	return le, nil
}

func renderJobLogs(ctx context.Context, logChannel <-chan string, logOutput *pterm.Logger, progress *pterm.ProgressbarPrinter) {
	lastMsg := ""
	for {
		select {
		case <-ctx.Done():
			return
		case line := <-logChannel:
			le, err := parseLogLine(line)
			if err != nil {
				logOutput.Debug(fmt.Sprintf("failed to unmarshal log entry: %s (was %q); ignoring...", err, line))
				continue
			}

			if le.MigrationStep > 0 {
				progress.Current = le.MigrationStep
				progress.UpdateTitle(le.Msg)
				continue
			}

			if lastMsg != le.Msg {
				args := logOutput.ArgsFromMap(le.extra)
				switch strings.ToLower(le.Level) {
				case "error":
					logOutput.Error(le.Msg, args)
				case "warn":
					logOutput.Warn(le.Msg, args)
				case "info":
					logOutput.Info(le.Msg, args)
				default:
					logOutput.Print(le.Msg, args)
				}
			}
			lastMsg = le.Msg
		}
	}
}
