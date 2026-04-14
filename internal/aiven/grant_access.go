package aiven

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	nais_kafka "github.com/nais/liberator/pkg/apis/kafka.nais.io/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

type GrantAccessResult struct {
	AlreadyAdded bool
}

func GrantAccessToTopic(ctx context.Context, namespace, topicName, environment string, newAcl nais_kafka.TopicACL) (*GrantAccessResult, error) {
	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(environment))

	if err := validateNamespace(ctx, client, namespace); err != nil {
		return nil, err
	}

	var topic nais_kafka.Topic
	if err := client.Get(ctx, ctrl.ObjectKey{Name: topicName, Namespace: namespace}, &topic); err != nil {
		return nil, fmt.Errorf("get topic: %w", err)
	}

	if checkIfAclInList(topic.Spec.ACL, newAcl) {
		return &GrantAccessResult{
			AlreadyAdded: true,
		}, nil
	}
	topic.Spec.ACL = append(topic.Spec.ACL, newAcl)

	if err := client.Update(ctx, &topic); err != nil {
		return nil, fmt.Errorf("update topic: %w", err)
	}

	return &GrantAccessResult{
		AlreadyAdded: false,
	}, nil
}

func checkIfAclInList(existing []nais_kafka.TopicACL, wanted nais_kafka.TopicACL) bool {
	for _, e := range existing {
		if e.Team == wanted.Team && e.Application == wanted.Application && e.Access == wanted.Access {
			return true
		}
	}
	return false
}

func ValidAclPermission(access string) error {
	switch access {
	case "read", "write", "readwrite":
		return nil
	default:
		return fmt.Errorf("invalid access type: %s (valid: read, write, readwrite)", access)
	}
}
