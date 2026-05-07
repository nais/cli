package flag

import (
	"github.com/nais/cli/internal/flags"
)

type (
	DebugSticky struct {
		*flags.GlobalFlags
		Copy bool `name:"copy" usage:"Create a copy of the pod with a debug container. The original pod remains running and unaffected."`
	}
)

type Debug struct {
	*DebugSticky
	ByPod bool `name:"by-pod" short:"b" usage:"Attach to a specific |BY-POD| in a workload."`
}

type DebugTidy struct {
	*DebugSticky
}
