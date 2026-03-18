package main

import (
	"fmt"
	"os"

	"github.com/nais/cli/internal/application"
)

func main() {
	app, _, err := application.New(os.Stdout)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := app.GenerateDocs(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Documentation generated successfully")
}
