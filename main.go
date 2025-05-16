package main

import (
	"context"
	"os"

	"github.com/nais/cli/internal/cli"
)

func main() {
	if err := cli.Run(context.Background()); err != nil {
		os.Exit(1)
	}
}
