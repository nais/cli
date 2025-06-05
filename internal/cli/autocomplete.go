package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func autocomplete(autoCompleteFunc AutoCompleteFunc, autoCompleteFilesExtensions []string) cobra.CompletionFunc {
	if len(autoCompleteFilesExtensions) > 0 {
		return autocompleteFiles(autoCompleteFilesExtensions)
	}

	if autoCompleteFunc != nil {
		return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			completions, activeHelp := autoCompleteFunc(cmd.Context(), args, toComplete)
			if activeHelp != "" {
				completions = cobra.AppendActiveHelp(completions, activeHelp)
			}
			return completions, cobra.ShellCompDirectiveDefault
		}
	}

	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	}
}

func autocompleteFiles(ext []string) cobra.CompletionFunc {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		helpSuffix := ""
		if num := len(ext); num > 0 {
			formatted := make([]string, num)
			for i, e := range ext {
				formatted[i] = "*." + e
			}
			helpSuffix = " (" + strings.Join(formatted[:num-1], ", ") + " or " + formatted[num-1] + ")"
		}

		ext = cobra.AppendActiveHelp(ext, fmt.Sprintf("Please choose one or more files%s.", helpSuffix))
		return ext, cobra.ShellCompDirectiveFilterFileExt
	}
}
