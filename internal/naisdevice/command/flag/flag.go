package flag

import "github.com/nais/cli/internal/root"

type Status struct {
	*root.Flags
	Quiet  bool
	Output string
}
