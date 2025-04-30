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
		Usage:       "The kubernetes `NAMESPACE` to use",
		DefaultText: "The namespace from your current kubeconfig context",
	}
}

func noWaitFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:  "no-wait",
		Usage: "Do not wait for the job to complete",
	}
}

func dryRunFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:  "dry-run",
		Usage: "Perform a dry run of the migration setup, without actually starting the migration",
	}
}
