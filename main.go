package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nais/cli/internal/cli"
)

func main() {
	if err := cli.Run(context.Background()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
