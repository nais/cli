package client

import (
	"k8s.io/client-go/rest"
)

type AivenClient struct {
	restClient rest.Interface
}

func (c *AivenClient) Aiven(namespace string) *aivenApplicationClient {
	return &aivenApplicationClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}
