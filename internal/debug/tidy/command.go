package tidy

import (
	"fmt"

	"github.com/nais/cli/v2/internal/debug"
	"github.com/nais/cli/v2/internal/debug/command/flag"
)

func Run(workloadName string, flags *flag.DebugTidy) error {
	clientSet, err := debug.SetupClient(flags.DebugSticky, flags.Context)
	if err != nil {
		return err
	}

	dg := debug.Setup(clientSet, flags.DebugSticky, workloadName, "", false)
	if err := dg.Tidy(); err != nil {
		return fmt.Errorf("debugging instance: %w", err)
	}

	return nil
}
