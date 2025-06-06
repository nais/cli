package cli

import (
	"context"
)

type (
	// RunFunc is a function that will be executed when the command is run.
	//
	// The args passed to this function is the arguments added to the command using the WithArgs option, in the same
	// order.
	RunFunc func(ctx context.Context, out Output, args []string) error

	// ValidateFunc is a function that will be executed before the command's RunFunc is executed.
	//
	// The args passed to this function is the arguments added to the command using the WithArgs option, in the same
	// order.
	ValidateFunc func(ctx context.Context, args []string) error

	// AutoCompleteFunc is a function that will be executed to provide auto-completion suggestions for the command.
	//
	// The args passed to this function is the arguments added to the command using the WithArgs option, in the same
	// order. toComplete is the current input that the user is typing, and it can be used to filter the suggestions.
	// The first return value is a slice of strings that will be used as suggestions, and the second return value is a
	// string that will be used as active help text in the shell while performing auto-complete.
	AutoCompleteFunc func(ctx context.Context, args []string, toComplete string) (completions []string, activeHelp string)
)
