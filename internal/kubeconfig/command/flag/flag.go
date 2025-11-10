package flag

import "github.com/nais/cli/internal/flags"

type Kubeconfig struct {
	*flags.GlobalFlags
	Exclude   []string `name:"exclude" short:"e" usage:"Exclude |CLUSTER| from kubeconfig. Can be repeated."`
	Overwrite bool     `name:"overwrite" short:"o" usage:"Overwrite existing kubeconfig entries if conflicts are found."`
	Clear     bool     `name:"clear" short:"c" usage:"Clear existing kubeconfig."`
}
