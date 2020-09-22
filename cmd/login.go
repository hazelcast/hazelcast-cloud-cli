package cmd

import (
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/spf13/cobra"
	"strings"
)

// Flags
var (
	apiKey    string
	apiSecret string
)

// Prompts
var (
	apiKeyQuestion = &survey.Password{
		Message: "Api Key:",
	}

	apiSecretQuestion = &survey.Password{
		Message: "Api Secret:",
	}
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:     "login",
	Aliases: []string{"login"},
	Short:   "This command logins you to Hazelcast Cloud with api-key and api-secret.",
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "- "
		})

		apiKeyQuestionErr := survey.AskOne(apiKeyQuestion, &apiKey, icons)

		if apiKeyQuestionErr != nil {
			return apiKeyQuestionErr
		}

		if len(strings.TrimSpace(apiKey)) == 0 {
			return errors.New("You must specify a valid api key to log in")
		}

		apiSecretQuestionErr := survey.AskOne(apiSecretQuestion, &apiSecret, icons)

		if apiSecretQuestionErr != nil {
			return apiSecretQuestionErr
		}

		if len(strings.TrimSpace(apiSecret)) == 0 {
			return errors.New("You must specify a valid api secret to log in")
		}

		configService := internal.NewConfigService()
		configService.Set(internal.ApiKey, apiKey)
		configService.Set(internal.ApiSecret, apiSecret)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
