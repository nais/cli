package main

import (
	"context"
	"os"

	"github.com/nais/cli/internal/cli"
)

func main() {
	if err := cli.Run(context.Background()); err != nil {
		// TODO: differentiate between cobra errors and internal errors
		// fmt.Println(err)
		os.Exit(1)
	}
}
