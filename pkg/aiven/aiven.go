package aiven

import (
	"fmt"
	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	"github.com/nais/liberator/pkg/namegen"
	aivenclient "github.com/nais/nais-d/client"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (
	AivenApiVersion          = "aiven.nais.io/v1"
	AivenKind                = "aivenApplication"
	DefaultProtected         = true
	MaxServiceUserNameLength = 64
)

type AivenConfiguration struct {
	client *kubernetes.Clientset
	AivenProperties
}

type AivenProperties struct {
	Username   string
	Namespace  string
	Pool       string
	Dest       string
	SecretName string
	Expiry     int
}

func SetupAivenConfiguration(client *kubernetes.Clientset, properties AivenProperties) *AivenConfiguration {
	return &AivenConfiguration{
		client,
		properties,
	}
}

func (a *AivenConfiguration) GenerateApplication() error {
	client, err := aivenclient.NewForConfig()
	if err != nil {
		panic(err)
	}

	timeStamp := time.Now().AddDate(0, 0, a.Expiry).Format(time.RFC3339)
	createApp := *a.CreateAivenApplication(timeStamp, a.SecretName)

	update := true
	existingAivenApp, err := client.Aiven(a.Namespace).Get(a.Username, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			createdApp, err := client.Aiven(a.Namespace).Create(&createApp)
			if err != nil {
				return err
			}
			update = false
			fmt.Printf("app: %s created\n", createdApp.Name)
		} else {
			return err
		}
	}

	if update {
		err := CopyMeta(existingAivenApp, &createApp)
		if err != nil {
			return err
		}
		createApp.Spec.ExpiresAt = existingAivenApp.Spec.ExpiresAt
		updatedApp, err := client.Aiven(a.Namespace).Update(&createApp)
		if err != nil {
			return err
		}
		fmt.Printf("aivenApp %s configured\n", updatedApp.Name)
	}

	fmt.Printf("To get secret and config run cmd --> 'nais-d aiven get %s %s -c kcat'", createApp.Spec.SecretName, a.Namespace)
	return nil
}

func (c AivenConfiguration) CreateAivenApplication(timeStamp, secretName string) *aiven_nais_io_v1.AivenApplication {
	app := &aiven_nais_io_v1.AivenApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       AivenKind,
			APIVersion: AivenApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Username,
			Namespace: c.Namespace,
		},
		Spec: aiven_nais_io_v1.AivenApplicationSpec{
			SecretName: "",
			Protected:  DefaultProtected,
			ExpiresAt:  timeStamp,
			Kafka:      &aiven_nais_io_v1.KafkaSpec{Pool: c.Pool},
		},
	}
	err := SetSecretName(app, secretName)
	if err != nil {
		return nil
	}
	return app
}

func SetSecretName(aivenApp *aiven_nais_io_v1.AivenApplication, secretName string) error {
	if secretName != "" {
		aivenApp.Spec.SecretName = secretName
	} else {
		newSecretName, err := setSecretName(aivenApp)
		if err != nil {
			return fmt.Errorf("could not create secretName: %s", err)
		}
		aivenApp.Spec.SecretName = newSecretName
	}
	return nil
}

func setSecretName(aivenApp *aiven_nais_io_v1.AivenApplication) (string, error) {
	return namegen.ShortName(SecretNamePrefix(aivenApp.Namespace, aivenApp.Name), MaxServiceUserNameLength)
}

func SecretNamePrefix(username, team string) string {
	return fmt.Sprintf("%s-%s", team, username)
}

// CopyMeta copies resource metadata from one resource to another.
// used when updating existing resources in the cluster.
func CopyMeta(src, dst runtime.Object) error {
	srcacc, err := meta.Accessor(src)
	if err != nil {
		return err
	}

	dstacc, err := meta.Accessor(dst)
	if err != nil {
		return err
	}

	dstacc.SetResourceVersion(srcacc.GetResourceVersion())
	dstacc.SetUID(srcacc.GetUID())
	dstacc.SetSelfLink(srcacc.GetSelfLink())

	return err
}
