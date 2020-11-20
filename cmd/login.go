package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:     "login",
	Aliases: []string{"login"},
	Short:   "This command logins you to Hazelcast Cloud with api-key and api-secret.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("API Key: ")
		apiKey, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
		fmt.Printf("\r\033[K")
		fmt.Print("API Secret: ")
		apiSecret, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
		fmt.Printf("\r\033[K")
		apiKeyString := strings.TrimSpace(string(apiKey))
		apiSecretString := strings.TrimSpace(string(apiSecret))

		loginResult, response, loginErr := internal.Login(apiKeyString, apiSecretString)
		internal.Validate(loginResult, response, loginErr)
		if loginErr == nil {
			configService := internal.NewConfigService()
			configService.Set(internal.ApiKey, apiKeyString)
			configService.Set(internal.ApiSecret, apiSecretString)
			color.Green("You have successfully logged into Hazelcast Cloud.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
