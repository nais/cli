package aiven

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	nais_kafka "github.com/nais/liberator/pkg/apis/kafka.nais.io/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

type GrantAccessResult struct {
	Added     bool
	Namespace string
	Name      string
	Kind      string
	Detail    string
}

func GrantAccessToTopic(ctx context.Context, namespace, topicName string, acl nais_kafka.TopicACL) (*GrantAccessResult, error) {
	client := k8s.SetupControllerRuntimeClient()

	if err := validateNamespace(ctx, client, namespace); err != nil {
		return nil, err
	}

	var topic nais_kafka.Topic
	if err := client.Get(ctx, ctrl.ObjectKey{Name: topicName, Namespace: namespace}, &topic); err != nil {
		return nil, fmt.Errorf("validate topic: %w", err)
	}

	// Default to read access if not specified
	if acl.Access == "" {
		acl.Access = "read"
	}

	newACLs, added := ensureTopicACL(topic.Spec.ACL, acl)
	if !added {
		return &GrantAccessResult{
			Added:     false,
			Namespace: namespace,
			Name:      topic.Name,
			Kind:      topic.Kind,
			Detail:    acl.Access,
		}, nil
	}

	topic.Spec.ACL = newACLs

	if err := client.Update(ctx, &topic); err != nil {
		return nil, fmt.Errorf("update topic: %w", err)
	}

	return &GrantAccessResult{
		Added:     true,
		Namespace: namespace,
		Name:      topic.Name,
		Kind:      topic.Kind,
		Detail:    acl.Access,
	}, nil
}

func GrantAccessToStream(ctx context.Context, namespace, streamName, userName string) (*GrantAccessResult, error) {
	client := k8s.SetupControllerRuntimeClient()

	if err := validateNamespace(ctx, client, namespace); err != nil {
		return nil, err
	}

	var stream nais_kafka.Stream
	if err := client.Get(ctx, ctrl.ObjectKey{Name: streamName, Namespace: namespace}, &stream); err != nil {
		return nil, fmt.Errorf("validate stream: %w", err)
	}

	newUsers, added := ensureAdditionalStreamUser(stream.Spec.AdditionalUsers, userName)
	if !added {
		return &GrantAccessResult{
			Added:     false,
			Namespace: namespace,
			Name:      stream.Name,
			Kind:      stream.Kind,
			Detail:    userName,
		}, nil
	}

	stream.Spec.AdditionalUsers = newUsers

	if err := client.Update(ctx, &stream); err != nil {
		return nil, fmt.Errorf("update stream: %w", err)
	}

	return &GrantAccessResult{
		Added:     true,
		Namespace: namespace,
		Name:      stream.Name,
		Kind:      stream.Kind,
		Detail:    userName,
	}, nil
}

func ensureTopicACL(existing []nais_kafka.TopicACL, wanted nais_kafka.TopicACL) ([]nais_kafka.TopicACL, bool) {
	for _, e := range existing {
		if e.Team == wanted.Team && e.Application == wanted.Application && e.Access == wanted.Access {
			return existing, false
		}
	}
	return append(existing, wanted), true
}

func ensureAdditionalStreamUser(existing []nais_kafka.AdditionalStreamUser, userName string) ([]nais_kafka.AdditionalStreamUser, bool) {
	for _, u := range existing {
		if u.Username == userName {
			return existing, false
		}
	}
	return append(existing, nais_kafka.AdditionalStreamUser{Username: userName}), true
}
