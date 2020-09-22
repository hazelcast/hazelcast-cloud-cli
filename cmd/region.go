package cmd

import (
	"context"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

var regionCloudProvider string

var regionCmd = &cobra.Command{
	Use:     "region",
	Aliases: []string{"r"},
	Short:   "This command used to collect supported cloud providers regions related information.",
}

var regionListCmd = &cobra.Command{
	Use:     "list",
	Short:   "This command lists available regions for Hazelcast Enterprise on selected cloud provider.",
	Example: "hzcloud region list --cloud-provider=azure",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		regions := internal.Validate(client.Region.List(context.Background(), &models.RegionInput{
			CloudProvider: regionCloudProvider,
		})).(*[]models.Region)
		header := table.Row{"#", "Name", "Available in Starter", "Available in Enterprise"}
		var rows []table.Row
		for k, cloudProvider := range *regions {
			rows = append(rows, table.Row{k + 1, cloudProvider.Name, cloudProvider.IsEnabledForStarter, cloudProvider.IsEnabledForEnterprise})
		}
		util.Print(util.PrintRequest{
			Data:       regions,
			Header:     header,
			Rows:       rows,
			PrintStyle: util.PrintStyle(outputStyle),
		})
	},
}

func init() {
	rootCmd.AddCommand(regionCmd)
	regionCmd.AddCommand(regionListCmd)
	regionListCmd.Flags().StringVar(&regionCloudProvider, "cloud-provider", "", "name of the cloud provider")
	err := regionListCmd.MarkFlagRequired("cloud-provider")
	if err != nil {
		panic(err)
	}
}
