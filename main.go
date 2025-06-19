package main

import (
	"context"
	"os"

	"github.com/nais/cli/internal/application"
)

func main() {
	if err := application.Run(context.Background(), os.Stdout); err != nil {
		os.Exit(1)
	}
}
