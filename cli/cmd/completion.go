package cmd

import (
	"os"

	"github.com/ecadlabs/auth/cli/auth"
	"github.com/spf13/cobra"
)

func NewCompletionCommand(client *auth.Client, rootCmd *cobra.Command) *cobra.Command {
	var completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Generates bash completion scripts",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
	}
	return completionCmd
}
