package cmd

import (
	"context"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

var instanceTypeCloudProvider string

var instanceTypeCmd = &cobra.Command{
	Use:   "instance-type",
	Aliases: []string{"it"},
	Short: "This command used to collect supported instance types for different cloud providers.",
}

var instanceTypeListCmd = &cobra.Command{
	Use:   "list",
	Short: "This command lists instance types that Hazelcast Enterprise supports.",
	Example: "hzcloud instance-type list",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		instanceTypes := internal.Validate(client.InstanceType.List(context.Background(), &models.InstanceTypeRequest{
			CloudProvider: instanceTypeCloudProvider,
		})).(*[]models.InstanceType)
		header := table.Row{"#", "Name", "Total Memory (GiB)"}
		var rows []table.Row
		for k, instanceType := range *instanceTypes {
			rows = append(rows, table.Row{k + 1, instanceType.Name, instanceType.TotalMemory})
		}
		util.Print(util.PrintRequest{
			Data:       instanceTypes,
			Header:     header,
			Rows:       rows,
			PrintStyle: util.PrintStyle(outputStyle),
		})
	},
}

func init() {
	rootCmd.AddCommand(instanceTypeCmd)
	instanceTypeCmd.AddCommand(instanceTypeListCmd)
	instanceTypeListCmd.Flags().StringVar(&instanceTypeCloudProvider, "cloud-provider", "", "name of the cloud provider")
	err := instanceTypeListCmd.MarkFlagRequired("cloud-provider")
	if err != nil {
		panic(err)
	}
}
