package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nais/cli/internal/application"
	"github.com/nais/cli/internal/metric"
)

func main() {
	if metric.RunUploader() {
		return
	}
	if err := application.Run(context.Background(), os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
