package client

import (
	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type AivenApplicationInterface interface {
	List(opts metav1.ListOptions) (*aiven_nais_io_v1.AivenApplicationList, error)
	Get(name string, options metav1.GetOptions) (*aiven_nais_io_v1.AivenApplication, error)
	Create(*aiven_nais_io_v1.AivenApplication) (*aiven_nais_io_v1.AivenApplication, error)
}

type aivenApplicationClient struct {
	restClient rest.Interface
	ns         string
}

type AivenInterface interface {
	Aiven(namespace string) AivenApplicationInterface
}

func NewForConfig() (*AivenClient, error) {
	c := GetConfig()
	err := AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}

	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: GroupName, Version: GroupVersion}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &AivenClient{restClient: client}, nil
}

func (c *aivenApplicationClient) Get(name string, opts metav1.GetOptions) (*aiven_nais_io_v1.AivenApplication, error) {
	result := aiven_nais_io_v1.AivenApplication{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("aivenapplications").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)

	return &result, err
}

func (c *aivenApplicationClient) Create(aivenApp *aiven_nais_io_v1.AivenApplication) (*aiven_nais_io_v1.AivenApplication, error) {
	result := aiven_nais_io_v1.AivenApplication{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("aivenapplications").
		Body(aivenApp).
		Do().
		Into(&result)

	return &result, err
}

func (c *aivenApplicationClient) Update(aivenApp *aiven_nais_io_v1.AivenApplication) (*aiven_nais_io_v1.AivenApplication, error) {
	result := aiven_nais_io_v1.AivenApplication{}
	err := c.restClient.
		Put().
		Namespace(c.ns).
		Resource("aivenapplications").
		Name(aivenApp.Name).
		Body(aivenApp).
		Do().
		Into(&result)

	return &result, err
}
