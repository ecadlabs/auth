package cmd

import (
	"fmt"

	"github.com/ecadlabs/auth/cli/auth"
	"github.com/spf13/cobra"
)

func NewGetCommand(client *auth.Client) *cobra.Command {
	get := &cobra.Command{
		Use:   "get [key]",
		Short: "Get an api key",
	}

	token := &cobra.Command{
		Use:   "token [role]",
		Short: "Get api token",
		Args:  cobra.MinimumNArgs(1),
		Run: func(command *cobra.Command, args []string) {
			user := command.Flag("user").Value.String()
			result, err := client.GetApiKeyToken(user, args[0])
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("%s", auth.JSONifyWhatever(result))
		},
	}

	token.PersistentFlags().String("user", "", "ID of the user")
	get.AddCommand(token)

	return get
}
