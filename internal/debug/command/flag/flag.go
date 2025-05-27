package flag

import "github.com/nais/cli/internal/root"

type Debug struct {
	*root.Flags
	Context   string
	Namespace string
	Copy      bool
	ByPod     bool
}

type DebugTidy struct {
	*Debug
}
