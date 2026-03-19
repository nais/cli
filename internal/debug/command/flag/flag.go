package flag

import (
	"fmt"

	"github.com/nais/cli/internal/flags"
)

type (
	Environment string
	DebugSticky struct {
		*flags.GlobalFlags
		Copy        bool        `name:"copy" usage:"Create a copy of the pod with a debug container. The original pod remains running and unaffected."`
		Environment Environment `name:"environment" short:"e" usage:"The environment to use."`
		Context     string      `name:"context" short:"c" usage:"REMOVED, see --environment."`
		Namespace   string      `name:"namespace" short:"n" usage:"REMOVED, see --team."`
	}
)

func (d DebugSticky) UsesRemovedFlags() error {
	if d.Namespace != "" {
		return fmt.Errorf("the -n, --namespace flag is replaced with the -t, --team flag")
	}
	if d.Context != "" {
		return fmt.Errorf("the -c, --context flag is replaced with the -e, --environment flag")
	}
	return nil
}

type Debug struct {
	*DebugSticky
	ByPod bool `name:"by-pod" short:"b" usage:"Attach to a specific |BY-POD| in a workload."`
}

type DebugTidy struct {
	*DebugSticky
}
