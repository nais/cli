package flag

import "github.com/nais/cli/internal/root"

type Kubeconfig struct {
	*root.Flags
	Exclude   []string
	Overwrite bool
	Clear     bool
}
