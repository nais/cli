package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

type Migrator struct {
	client ctrl.Client
	cfg    Config
}

func NewMigrator(client ctrl.Client, cfg Config) *Migrator {
	return &Migrator{
		client,
		cfg,
	}
}

func (m *Migrator) Setup(ctx context.Context) error {
	fmt.Println("Resolving target instance config")
	err := m.cfg.Target.Resolve(ctx, m.client, m.cfg.AppName, m.cfg.Namespace)
	if err != nil {
		return err
	}
	fmt.Printf("Resolved target:\n%s\n", m.cfg.Target.String())

	fmt.Println("Resolving source instance config")
	err = m.cfg.Source.Resolve(ctx, m.client, m.cfg.AppName, m.cfg.Namespace)
	if err != nil {
		return err
	}
	fmt.Printf("Resolved source:\n%s\n", m.cfg.Source.String())

	fmt.Println("Creating ConfigMap")
	cfgMap := m.cfg.CreateConfigMap()
	err = m.client.Create(ctx, &cfgMap)
	if err != nil {
		return fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	imageTag, err := getLatestImageTag()
	if err != nil {
		return fmt.Errorf("failed to get latest image tag for cloudsql-migrator: %w", err)
	}

	job := createJob(m.cfg.AppName, m.cfg.Namespace, imageTag)
	err = m.client.Create(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to create NaisJob: %w", err)
	}

	fmt.Println("TODO: Do more stuff")
	return fmt.Errorf("TODO: Do more stuff")
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

func createJob(appName, namespace, imageTag string) *nais_io_v1.Naisjob {
	migrationName := fmt.Sprintf("%s-migrate-config", appName)
	return &nais_io_v1.Naisjob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      migrationName,
			Namespace: namespace,
		},
		Spec: nais_io_v1.NaisjobSpec{
			Command: []string{"/setup"},
			EnvFrom: []nais_io_v1.EnvFrom{{
				ConfigMap: migrationName,
			}},
			GCP: &nais_io_v1.GCP{
				Permissions: []nais_io_v1.CloudIAMPermission{{
					Role: "roles/cloudsql.admin",
					Resource: nais_io_v1.CloudIAMResource{
						APIVersion: "resourcemanager.cnrm.cloud.google.com/v1beta1",
						Kind:       "Project",
					},
				}},
			},
			Image: fmt.Sprintf("europe-north1-docker.pkg.dev/nais-io/nais/images/cloudsql-migrator:%s", imageTag),
		},
	}
}
