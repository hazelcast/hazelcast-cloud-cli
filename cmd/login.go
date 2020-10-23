package cmd

import (
	"bufio"
	"fmt"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"syscall"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:     "login",
	Aliases: []string{"login"},
	Short:   "This command logins you to Hazelcast Cloud with api-key and api-secret.",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Api Key: ")
		apiKey, _ := reader.ReadString('\n')
		fmt.Print("Api Secret: ")
		apiSecret, _ := terminal.ReadPassword(syscall.Stdin)
		configService := internal.NewConfigService()
		configService.Set(internal.ApiKey, strings.TrimSpace(apiKey))
		configService.Set(internal.ApiSecret, strings.TrimSpace(string(apiSecret)))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
