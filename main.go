package main

import (
	"fmt"
	"os"

	cli "github.com/nais/cli/internal/cli2"
)

func main() {
	if err := cli.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
