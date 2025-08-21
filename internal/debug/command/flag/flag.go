package flag

import (
	"time"

	"github.com/nais/cli/internal/root"
)

type Context string

type Debug struct {
	*root.Flags
	Context   Context       `short:"c" usage:"The kubeconfig |context| to use. Defaults to current context."`
	Namespace string        `short:"n" usage:"The kubernetes |namespace| to use. Defaults to current namespace."`
	Copy      bool          `usage:"Create a copy of the pod with a debug container. The original pod remains running and unaffected."`
	TTL       time.Duration `usage:"|Duration| the debug pod remains after exit. Only has effect when --copy is specified."`
	Timeout   time.Duration `usage:"|Duration| to wait for each remote interaction this command does. Usually the default is sufficient."`
}
