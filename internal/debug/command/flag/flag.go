package flag

import (
	"time"

	"github.com/nais/cli/internal/root"
)

type Context string

type Debug struct {
	*root.Flags
	Context   Context       `name:"context" short:"c" usage:"The kubeconfig |context| to use. Defaults to current context."`
	Namespace string        `name:"namespace" short:"n" usage:"The kubernetes |namespace| to use. Defaults to current namespace."`
	Copy      bool          `name:"copy" usage:"Create a copy of the pod with a debug container. The original pod remains running and unaffected."`
	Ttl       time.Duration `name:"ttl" usage:"|Duration| the debug pod remains after exit. Only has effect when --copy is specified."`
}
