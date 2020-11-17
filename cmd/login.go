package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	hazelcastcloud "github.com/hazelcast/hazelcast-cloud-sdk-go"
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
		fmt.Print("- API Key: ")
		apiKey, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
		fmt.Printf("\r\033[K")
		fmt.Print("- API Secret: ")
		apiSecret, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
		fmt.Printf("\r\033[K")
		apiKeyString := strings.TrimSpace(string(apiKey))
		apiSecretString := strings.TrimSpace(string(apiSecret))

		apiUrl := os.Getenv("HZ_CLOUD_API_URL")
		var clientErr error
		if len(apiUrl) != 0 {
			_, _, clientErr = hazelcastcloud.NewFromCredentials(apiKeyString, apiSecretString,
				hazelcastcloud.OptionEndpoint(apiUrl))
		} else {
			_, _, clientErr = hazelcastcloud.NewFromCredentials(apiKeyString, apiSecretString)
		}
		internal.Validate(nil, nil, clientErr)
		if clientErr == nil {
			configService := internal.NewConfigService()
			configService.Set(internal.ApiKey, apiKeyString)
			configService.Set(internal.ApiSecret, apiSecretString)
			color.Green("Login successful.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
