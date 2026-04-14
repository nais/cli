package aiven

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

func validateNamespace(ctx context.Context, client ctrl.Client, name string) error {
	var namespace v1.Namespace
	err := client.Get(ctx, ctrl.ObjectKey{Name: name}, &namespace)
	if err != nil {
		return fmt.Errorf("get namespace: %w", err)
	}

	return nil
}
