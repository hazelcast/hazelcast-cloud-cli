package cmd

import (
	"context"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var cloudProviderCmd = &cobra.Command{
	Use:     "cloud-provider",
	Aliases: []string{"cp"},
	Short:   "This command used to collect supported cloud provider related information.",
}

var cloudProviderListCmd = &cobra.Command{
	Use:     "list",
	Short:   "This command lists a available cloud provider list that Hazelcast Cloud supports.",
	Example: "hzcloud cloud-provider list",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		cloudProvider := internal.Validate(client.CloudProvider.List(context.Background())).(*[]models.CloudProvider)
		header := table.Row{"#", "Name", "Available in Starter", "Available in Enterprise"}
		rows := []table.Row{}
		for k, cloudProvider := range *cloudProvider {
			rows = append(rows, table.Row{k + 1, cloudProvider.Name, cloudProvider.IsEnabledForStarter, cloudProvider.IsEnabledForEnterprise})
		}
		util.Print(util.PrintRequest{
			Data:       cloudProvider,
			Header:     header,
			Rows:       rows,
			PrintStyle: util.PrintStyle(outputStyle),
		})
	},
}

func init() {
	rootCmd.AddCommand(cloudProviderCmd)
	cloudProviderCmd.AddCommand(cloudProviderListCmd)
}
