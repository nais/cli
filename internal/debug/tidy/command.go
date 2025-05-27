package tidy

import (
	"fmt"

	"github.com/nais/cli/internal/debug"
	"github.com/nais/cli/internal/debug/command/flag"
)

func Run(workloadName string, flags *flag.DebugTidy) error {
	cfg := debug.MakeConfig(workloadName, flags.Debug)
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
