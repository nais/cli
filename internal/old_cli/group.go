package cli

import "github.com/spf13/cobra"

var authGroup = &cobra.Group{
	ID:    "auth",
	Title: "Authentication",
}
