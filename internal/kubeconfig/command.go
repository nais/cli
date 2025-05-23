package kubeconfig

import (
	"context"

	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/root"
)

type Flags struct {
	*root.Flags
	Exclude   []string
	Overwrite bool
	Clear     bool
	Verbose   bool
}

func Run(ctx context.Context, flags *Flags) error {
	email, err := gcp.GetActiveUserEmail(ctx)
	if err != nil {
		return err
	}

	return CreateKubeconfig(
		ctx,
		email,
		WithOverwriteData(flags.Overwrite),
		WithFromScratch(flags.Clear),
		WithExcludeClusters(flags.Exclude),
		WithOnpremClusters(true),
		WithVerboseLogging(flags.Verbose),
	)
}
