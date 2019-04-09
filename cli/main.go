package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/ecadlabs/auth/cli/auth"
	"github.com/ecadlabs/auth/cli/cmd"
	"github.com/ecadlabs/auth/cli/config"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func getUrl() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nApi Url: ")
	text, _ := reader.ReadString('\n')
	return strings.Trim(text, "\n")
}

func getUsername() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nEmail: ")
	text, _ := reader.ReadString('\n')
	return strings.Trim(text, "\n")
}

func getPassword() string {
	fmt.Printf("\nPassword: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("Failed to read password: %v", err)
	}
	fmt.Println("")
	return strings.Trim(string(bytePassword), "\n")
}

func createClient() (*auth.Client, error) {
	c, err := config.LoadYAML("./config.yaml")

	if err != nil {
		url := getUrl()
		client := auth.New(url)

		loginResult, err := client.Login(getUsername(), getPassword())

		if err != nil {
			return nil, err
		}

		c = &config.Config{
			URL:   url,
			Token: loginResult.Token,
		}

		err = config.Persist("./config.yaml", c)

		if err != nil {
			return nil, err
		}
	}
	client := auth.New(c.URL)
	client.SetToken(c.Token)
	return client, nil
}

func main() {
	var rootCmd = &cobra.Command{Use: "auth"}

	client, err := createClient()

	if err != nil {
		panic(err)
	}

	rootCmd.AddCommand(cmd.NewCompletionCommand(client, rootCmd))
	rootCmd.AddCommand(cmd.NewCreateCommand(client))
	rootCmd.AddCommand(cmd.NewAddCommand(client))
	rootCmd.AddCommand(cmd.NewGetCommand(client))
	rootCmd.Execute()
}
