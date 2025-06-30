package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nais/cli/v2/internal/application"
)

func main() {
	if err := application.Run(context.Background(), os.Stdout); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
