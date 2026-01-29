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
	Namespace    string
	Name         string
}

func GrantAccessToTopic(ctx context.Context, namespace, topicName string, newAcl nais_kafka.TopicACL) (*GrantAccessResult, error) {
	client := k8s.SetupControllerRuntimeClient()

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
			Namespace:    namespace,
			Name:         topicName,
		}, nil
	}
	topic.Spec.ACL = append(topic.Spec.ACL, newAcl)

	if err := client.Update(ctx, &topic); err != nil {
		return nil, fmt.Errorf("update topic: %w", err)
	}

	return &GrantAccessResult{
		AlreadyAdded: false,
		Namespace:    namespace,
		Name:         topicName,
	}, nil
}

func GrantAccessToStream(ctx context.Context, namespace, streamName, userName string) (*GrantAccessResult, error) {
	client := k8s.SetupControllerRuntimeClient()

	if err := validateNamespace(ctx, client, namespace); err != nil {
		return nil, err
	}

	var stream nais_kafka.Stream
	if err := client.Get(ctx, ctrl.ObjectKey{Name: streamName, Namespace: namespace}, &stream); err != nil {
		return nil, fmt.Errorf("get stream: %w", err)
	}

	if checkIfUserInList(stream.Spec.AdditionalUsers, userName) {
		return &GrantAccessResult{
			AlreadyAdded: true,
			Namespace:    namespace,
			Name:         streamName,
		}, nil
	}
	stream.Spec.AdditionalUsers = append(stream.Spec.AdditionalUsers, nais_kafka.AdditionalStreamUser{Username: userName})

	if err := client.Update(ctx, &stream); err != nil {
		return nil, fmt.Errorf("update stream: %w", err)
	}

	return &GrantAccessResult{
		AlreadyAdded: false,
		Namespace:    namespace,
		Name:         streamName,
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

func checkIfUserInList(existing []nais_kafka.AdditionalStreamUser, userName string) bool {
	for _, u := range existing {
		if u.Username == userName {
			return true
		}
	}
	return false
}
