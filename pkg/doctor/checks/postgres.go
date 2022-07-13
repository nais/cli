package checks

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/cloudsql-proxy/logging"
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"

	gk8sv1alpha1 "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/clients/generated/apis/k8s/v1alpha1"
	gsqlv1beta1 "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/clients/generated/apis/sql/v1beta1"
	"github.com/nais/cli/pkg/doctor"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Postgres struct {
	cfg *doctor.Config
}

func init() {
	doctor.AddCheck(&Postgres{})
}

func (s *Postgres) Name() string {
	return "postgres"
}

func (s *Postgres) Help() string {
	return "Check access to postgres instances using the application credentials."
}

func (s *Postgres) Ack() {}

func (s *Postgres) Check(ctx context.Context, cfg *doctor.Config) []error {
	s.cfg = cfg

	if cfg.Application.Spec.GCP == nil || cfg.Application.Spec.GCP.SqlInstances == nil {
		cfg.Log.Info("no postgres instances defined in the application spec")
		return []error{doctor.ErrSkip}
	}

	psql := cfg.Application.Spec.GCP.SqlInstances

	errs := []error{}
	for _, instance := range psql {
		if err := s.check(ctx, instance); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (s *Postgres) check(ctx context.Context, instance nais_io_v1.CloudSqlInstance) error {
	name := instance.Name
	if name == "" {
		name = instance.Databases[0].Name
		s.cfg.Log.Debug("no name on instance, falling back to first database name")
	}
	log := s.cfg.Log.WithField("instance", name)

	log.Info("checking postgres instance")

	sqlInstance, err := s.postgresqlGetSqlInstance(ctx, log, name)
	if err != nil {
		return err
	}

	if err := s.sqlInstanceReady(log, sqlInstance.Status.Conditions); err != nil {
		return err
	}

	if err := s.sqlUserExists(ctx, log, name); err != nil {
		return err
	}

	log.Debug("checking if credentials are correct")
	if err := s.ping(ctx, log, name, sqlInstance); err != nil {
		return err
	}

	return nil
}

func (s *Postgres) postgresqlGetSqlInstance(ctx context.Context, log *logrus.Entry, instance string) (*gsqlv1beta1.SQLInstance, error) {
	log.Debug("finding sqlinstance")
	kinstance, err := s.cfg.DynamicClient.Resource(schema.GroupVersionResource{
		Group:    "sql.cnrm.cloud.google.com",
		Version:  "v1beta1",
		Resource: "sqlinstances",
	}).Namespace(s.cfg.Application.Namespace).Get(ctx, instance, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, doctor.ErrorMsg(err, "PostgreSQL instance not found, try to redeploy the application")
		}
		return nil, err
	}

	ret := &gsqlv1beta1.SQLInstance{}
	return ret, runtime.DefaultUnstructuredConverter.FromUnstructured(kinstance.Object, ret)
}

func (s *Postgres) sqlInstanceReady(log *logrus.Entry, conditions []gk8sv1alpha1.Condition) error {
	log.Debug("checking if sqlinstance is ready")

	for _, condition := range conditions {
		if condition.Type == gk8sv1alpha1.ReadyConditionType {
			if condition.Status == corev1.ConditionTrue {
				return nil
			}
			return doctor.ErrorMsg(nil, "PostgreSQL instance is not ready")
		}
	}
	return nil
}

func (s *Postgres) sqlUserExists(ctx context.Context, log *logrus.Entry, name string) error {
	log.WithField("user", name).Debug("checking if user exists")
	_, err := s.cfg.DynamicClient.Resource(schema.GroupVersionResource{
		Group:    "sql.cnrm.cloud.google.com",
		Version:  "v1beta1",
		Resource: "sqlusers",
	}).Namespace(s.cfg.Application.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return doctor.ErrorMsg(err, "PostgreSQL user not found, try to redeploy the application")
		}
		return err
	}

	return nil
}

func (s *Postgres) ping(ctx context.Context, log *logrus.Entry, name string, instance *gsqlv1beta1.SQLInstance) error {
	secret, err := s.cfg.K8sClient.CoreV1().Secrets(s.cfg.Application.Namespace).Get(ctx, "google-sql-"+name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return doctor.ErrorMsg(err, "PostgreSQL credentials 'google-sql"+name+"' not found, try to redeploy the application")
		}
		return err
	}

	if secret.Data == nil {
		return doctor.ErrorMsg(nil, "PostgreSQL credentials empty, try to redeploy the application")
	}

	username := getSecretDataValue(secret, "_USERNAME")
	password := getSecretDataValue(secret, "_PASSWORD")
	dbName := getSecretDataValue(secret, "_DATABASE")
	host := instance.Status.ConnectionName
	conStr := fmt.Sprintf("host=%v user=%v dbname=%v password=%v sslmode=disable", host, username, dbName, password)

	logging.DisableLogging()
	log.WithFields(map[string]any{
		"host":     host,
		"username": username,
		"dbname":   dbName,
	}).Debug("pinging postgres instance")

	db, err := sql.Open("cloudsqlpostgres", conStr)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return doctor.ErrorMsg(doctor.ErrWarning, "Pinging PostgreSQL instance failed with error (You might need to run `gcloud auth login --update-adc`): "+err.Error())
	}

	return nil
}

func getSecretDataValue(secret *corev1.Secret, suffix string) string {
	for name, val := range secret.Data {
		if strings.HasSuffix(name, suffix) {
			return string(val)
		}
	}
	return ""
}
