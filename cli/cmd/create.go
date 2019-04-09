package cmd

import (
	"fmt"

	"github.com/ecadlabs/auth/cli/auth"
	"github.com/spf13/cobra"
)

func NewCreateCommand(client *auth.Client) *cobra.Command {
	create := &cobra.Command{
		Use:   "create",
		Short: "Create a user or a tenant",
	}

	createUser := &cobra.Command{
		Use:   "user",
		Short: "Create a user",
		Args:  cobra.MinimumNArgs(1),
		Run: func(command *cobra.Command, args []string) {
			userName := args[0]
			accountType := command.Flag("type").Value.String()
			email := command.Flag("email").Value.String()
			result, err := client.CreateUser(&auth.User{Name: userName, Email: email, Type: accountType})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("%s", auth.JSONifyWhatever(result))
		},
	}

	createService := &cobra.Command{
		Use:   "service",
		Short: "Create a service account and provision api key",
		Args:  cobra.MinimumNArgs(0),
		Run: func(command *cobra.Command, args []string) {
			tenant := command.Flag("tenant").Value.String()
			username := command.Flag("user").Value.String()
			role := command.Flag("role").Value.String()
			user, err := client.CreateUser(&auth.User{Name: username, Type: "service"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			err = client.CreateMembership(&auth.Membership{Role: role, UserID: user.ID, TenantID: tenant})
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			apiKey, err := client.CreateApiKey(user.ID, tenant)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			apiKeyToken, err := client.GetApiKeyToken(user.ID, apiKey.ID.String())
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			fmt.Printf("Service account ID: %s\n", user.ID)
			fmt.Printf("Tenant ID: %s\n", tenant)
			fmt.Printf("Api Key ID: %s\n", apiKey.ID)
			fmt.Printf("Token: %s\n", apiKeyToken.Token)
		},
	}

	createMembership := &cobra.Command{
		Use:   "membership",
		Short: "Create a membership",
		Args:  cobra.MinimumNArgs(0),
		Run: func(command *cobra.Command, args []string) {
			tenant := command.Flag("tenant").Value.String()
			user := command.Flag("user").Value.String()
			role := command.Flag("role").Value.String()
			err := client.CreateMembership(&auth.Membership{Role: role, UserID: user, TenantID: tenant})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("Membership created")
		},
	}

	createAPIKey := &cobra.Command{
		Use:   "api_key",
		Short: "Create an api key",
		Args:  cobra.MinimumNArgs(0),
		Run: func(command *cobra.Command, args []string) {
			user := command.Flag("user").Value.String()
			tenant := command.Flag("tenant").Value.String()
			result, err := client.CreateApiKey(user, tenant)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("%s", auth.JSONifyWhatever(result))
		},
	}

	createService.PersistentFlags().String("user", "", "Name of the user")
	createService.PersistentFlags().String("tenant", "", "Name of the user")
	createService.PersistentFlags().String("role", "", "Name of the user")

	createUser.PersistentFlags().String("name", "", "Name of the user")
	createUser.PersistentFlags().String("type", "regular", "Account type (service|regular)")
	createUser.PersistentFlags().String("email", "", "User email")

	createMembership.PersistentFlags().String("user", "", "Name of the user")
	createMembership.PersistentFlags().String("tenant", "", "Name of the user")
	createMembership.PersistentFlags().String("role", "", "Name of the user")

	createAPIKey.PersistentFlags().String("user", "", "Name of the user")
	createAPIKey.PersistentFlags().String("tenant", "", "Name of the user")

	create.AddCommand(createService)
	create.AddCommand(createUser)
	create.AddCommand(createMembership)
	create.AddCommand(createAPIKey)

	return create
}
