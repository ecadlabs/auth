package cmd

import (
	"fmt"

	"github.com/ecadlabs/auth/cli/auth"
	"github.com/spf13/cobra"
)

func NewAddCommand(client *auth.Client) *cobra.Command {
	add := &cobra.Command{
		Use:   "add",
		Short: "Add an ip or a role",
		Args:  cobra.MinimumNArgs(1),
	}

	role := &cobra.Command{
		Use:   "role",
		Short: "Add a role to a membership",
		Args:  cobra.MinimumNArgs(1),
		Run: func(command *cobra.Command, args []string) {
			fmt.Println(command.Flag("user").Value.String())
			fmt.Printf("%s", args)
		},
	}

	ip := &cobra.Command{
		Use:   "ip",
		Short: "Add an ip to a user",
		Args:  cobra.MinimumNArgs(1),
		Run: func(command *cobra.Command, args []string) {
			user := command.Flag("user").Value.String()
			result, err := client.AddIp(&auth.AddIpRequest{UserID: user, IPs: args})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("%s", auth.JSONifyWhatever(result))
		},
	}

	role.PersistentFlags().String("user", "", "Name of the user")
	ip.PersistentFlags().String("user", "", "Name of the user")
	role.PersistentFlags().String("tenant", "regular", "Account type (service|regular)")
	add.AddCommand(role)
	add.AddCommand(ip)

	return add
}
