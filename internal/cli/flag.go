package cli

import "github.com/urfave/cli/v3"

func contextFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "context",
		Aliases:     []string{"c"},
		Usage:       "The kubeconfig `CONTEXT` to use",
		DefaultText: "The current context in your kubeconfig",
	}
}

func byPodFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "by-pod",
		Aliases:     []string{"p"},
		Usage:       "Attach to a specific `BY-POD` in a workload",
		DefaultText: "The first pod in the workload",
	}
}

func copyFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "copy",
		Aliases:     []string{"cp"},
		Usage:       "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected",
		DefaultText: "Attach to the current 'live' pod",
	}
}

func namespaceFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "namespace",
		Aliases:     []string{"n"},
		Usage:       "The `NAMESPACE` to use",
		DefaultText: "The current namespace in your kubeconfig",
	}
}
