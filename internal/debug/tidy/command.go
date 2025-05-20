package tidy

import (
	"fmt"

	"github.com/nais/cli/internal/debug"
	"github.com/nais/cli/internal/root"
)

type Flags struct {
	*root.Flags
	Context   string
	Namespace string
	Copy      bool
}

func Run(workloadName string, flags *Flags) error {
	cfg := debug.MakeConfig(workloadName, &debug.Flags{
		Context:   flags.Context,
		Namespace: flags.Namespace,
		Copy:      flags.Copy,
	})
	clientSet, err := debug.SetupClient(cfg, flags.Context)
	if err != nil {
		return err
	}

	dg := debug.Setup(clientSet, cfg)
	if err := dg.Tidy(); err != nil {
		return fmt.Errorf("debugging instance: %w", err)
	}

	return nil
}
