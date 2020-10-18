package cmd

import (
	"context"
	"github.com/fatih/color"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/hazelcast/hazelcast-cloud-cli/service"
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

var gcpPeeringId string
var gcpNetworkName string
var gcpProjectId string

var gcpPeeringCmd = &cobra.Command{
	Use:     "gcp-peering",
	Short:   "This command allows you to make Google Cloud Platform VPC Peering related actions.",
	Aliases: []string{"gcp"},
}

var gcpPeeringCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "This command creates GCP VPC Peering between your own VPC and your Enterprise Hazelcast cluster vNet.",
	Example: "hzcloud gcp-peering create --cluster-id=1 --project-id=2 --network-name=3",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		peeringCreateErr := service.NewGcpPeeringService(client).Create(&service.GcpCustomerPeeringProperties{
			ClusterId:   enterpriseClusterId,
			ProjectId:   gcpProjectId,
			NetworkName: gcpNetworkName,
		})
		if peeringCreateErr != nil {
			color.Red("An error occurred. %s", peeringCreateErr)
		} else {
			color.Green("Peering successfully established.")
		}
	},
}

var gcpPeeringListCmd = &cobra.Command{
	Use:     "list",
	Short:   "This command lists GCP VPC peerings on your Enterprise Hazelcast cluster.",
	Example: "hzcloud gcp-peering list --cluster-id=1",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		peerings := internal.Validate(client.GcpPeering.List(context.Background(), &models.ListGcpPeeringsInput{
			ClusterId: enterpriseClusterId,
		})).(*[]models.GcpPeering)
		header := table.Row{"#", "Peering Id", "Project Id", "Network Name"}
		rows := []table.Row{}
		for k, peering := range *peerings {
			rows = append(rows, table.Row{k + 1, peering.Id, peering.ProjectId, peering.NetworkName})
		}
		util.Print(util.PrintRequest{
			Data:       peerings,
			Header:     header,
			Rows:       rows,
			PrintStyle: util.PrintStyle(outputStyle),
		})
	},
}

var gcpPeeringDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "This command deletes GCP VPC peering from your Enterprise Hazelcast cluster.",
	Example: "hzcloud gcp-peering delete --peering-id=1",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		_ = internal.Validate(client.GcpPeering.Delete(context.Background(), &models.DeleteGcpPeeringInput{
			Id: gcpPeeringId,
		})).(*models.Result)
		color.Blue("Peering %s deleted.", gcpPeeringId)
	},
}

func init() {
	rootCmd.AddCommand(gcpPeeringCmd)
	gcpPeeringCmd.AddCommand(gcpPeeringCreateCmd)
	gcpPeeringCmd.AddCommand(gcpPeeringListCmd)
	gcpPeeringCmd.AddCommand(gcpPeeringDeleteCmd)

	gcpPeeringListCmd.Flags().StringVar(&enterpriseClusterId, "cluster-id", "", "id of the cluster")
	_ = gcpPeeringListCmd.MarkFlagRequired("cluster-id")

	gcpPeeringDeleteCmd.Flags().StringVar(&gcpPeeringId, "peering-id", "", "id of the peering")
	_ = gcpPeeringDeleteCmd.MarkFlagRequired("peering-id")

	gcpPeeringCreateCmd.Flags().StringVar(&enterpriseClusterId, "cluster-id", "", "id of the cluster")
	_ = gcpPeeringCreateCmd.MarkFlagRequired("cluster-id")

	gcpPeeringCreateCmd.Flags().StringVar(&gcpNetworkName, "network-name", "", "id of the cluster")
	_ = gcpPeeringCreateCmd.MarkFlagRequired("network-name")

	gcpPeeringCreateCmd.Flags().StringVar(&gcpProjectId, "project-id", "", "id of the cluster")
	_ = gcpPeeringCreateCmd.MarkFlagRequired("project-id")

}
