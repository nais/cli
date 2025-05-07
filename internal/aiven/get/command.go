package get

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
)

type Arguments struct {
	SecretName string
	Namespace  string
}

func Run(ctx context.Context, service aiven_services.Service, args Arguments) error {
	if err := aiven.ExtractAndGenerateConfig(ctx, service, args.SecretName, args.Namespace); err != nil {
		return fmt.Errorf("retrieve secret and generating config: %w", err)
	}

	return nil
}
