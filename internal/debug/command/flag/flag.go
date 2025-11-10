package flag

import "github.com/nais/cli/internal/flags"

type (
	Context     string
	DebugSticky struct {
		*flags.GlobalFlags
		Context   Context `name:"context" short:"c" usage:"The kubeconfig |CONTEXT| to use. Defaults to current context."`
		Namespace string  `name:"namespace" short:"n" usage:"The kubernetes |NAMESPACE| to use. Defaults to current namespace."`
		Copy      bool    `name:"copy" usage:"Create a copy of the pod with a debug container. The original pod remains running and unaffected."`
	}
)

type Debug struct {
	*DebugSticky
	ByPod bool `name:"by-pod" short:"b" usage:"Attach to a specific |BY-POD| in a workload."`
}

type DebugTidy struct {
	*DebugSticky
}
