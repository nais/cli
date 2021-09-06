package main

import "github.com/nais/debuk/cmd"

var (
	// VERSION is set during build
	VERSION = "v0.1"
)

func main() {
	cmd.Execute(VERSION)
}
