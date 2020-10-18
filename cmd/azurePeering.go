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

var azurePeeringId string
var azureTenantId string
var azureSubscriptionId string
var azureResourceGroupName string
var azureVnetName string

var azurePeeringCmd = &cobra.Command{
	Use:     "azure-peering",
	Short:   "This command allows you to make Azure Virtual Network Peering related actions.",
	Aliases: []string{"azure"},
}

var azurePeeringCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "This command creates Azure vNet Peering between your own vNet and your Enterprise Hazelcast cluster vNet.",
	Example: "hzcloud azure-peering create --cluster-id=1 --project-id=2 --network-name=3",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		indicator := util.NewLoadingIndicator("Azure Peering starting...", 100)
		indicator.Start()
		azurePeeringService := service.NewAzurePeeringService(client,&service.AzureCustomerPeeringProperties{
			ClusterId:         enterpriseClusterId,
			TenantId:          azureTenantId,
			SubscriptionId:    azureSubscriptionId,
			ResourceGroupName: azureResourceGroupName,
			VnetName:          azureVnetName,
		})
		peeringCreateErr := azurePeeringService.Create(indicator)
		indicator.Stop()
		if peeringCreateErr != nil {
			color.Red("An error occurred. %s", peeringCreateErr)
		} else {
			color.Green("Peering successfully established.")
		}
	},
}

var azurePeeringListCmd = &cobra.Command{
	Use:     "list",
	Short:   "This command lists Azure vNet peerings on your Enterprise Hazelcast cluster.",
	Example: "hzcloud azure-peering list --cluster-id=1",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		peerings := internal.Validate(client.AzurePeering.List(context.Background(), &models.ListAzurePeeringsInput{
			ClusterId: enterpriseClusterId,
		})).(*[]models.AzurePeering)
		header := table.Row{"#", "Peering Id"}
		rows := []table.Row{}
		for k, peering := range *peerings {
			rows = append(rows, table.Row{k + 1, peering.Id})
		}
		util.Print(util.PrintRequest{
			Data:       peerings,
			Header:     header,
			Rows:       rows,
			PrintStyle: util.PrintStyle(outputStyle),
		})
	},
}

var azurePeeringDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "This command deletes Azure vNet peering from your Enterprise Hazelcast cluster.",
	Example: "hzcloud azure-peering delete --peering-id=1",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		_ = internal.Validate(client.AzurePeering.Delete(context.Background(), &models.DeleteAzurePeeringInput{
			Id: azurePeeringId,
		})).(*models.Result)
		color.Blue("Peering %s deleted.", azurePeeringId)
	},
}

func init() {
	rootCmd.AddCommand(azurePeeringCmd)
	azurePeeringCmd.AddCommand(azurePeeringCreateCmd)
	azurePeeringCmd.AddCommand(azurePeeringListCmd)
	azurePeeringCmd.AddCommand(azurePeeringDeleteCmd)

	azurePeeringListCmd.Flags().StringVar(&enterpriseClusterId, "cluster-id", "", "id of the cluster")
	_ = azurePeeringListCmd.MarkFlagRequired("cluster-id")

	azurePeeringDeleteCmd.Flags().StringVar(&azurePeeringId, "peering-id", "", "id of the peering")
	_ = azurePeeringDeleteCmd.MarkFlagRequired("peering-id")

	azurePeeringCreateCmd.Flags().StringVar(&enterpriseClusterId, "cluster-id", "", "id of the cluster")
	_ = azurePeeringCreateCmd.MarkFlagRequired("cluster-id")
	azurePeeringCreateCmd.Flags().StringVar(&azureTenantId, "tenant-id", "", "id of the azure tenant")
	_ = azurePeeringCreateCmd.MarkFlagRequired("tenant-id")
	azurePeeringCreateCmd.Flags().StringVar(&azureResourceGroupName, "resource-group", "", "name of the azure resource group")
	_ = azurePeeringCreateCmd.MarkFlagRequired("resource-group")
	azurePeeringCreateCmd.Flags().StringVar(&azureSubscriptionId, "subscription-id", "", "id of the azure subscription")
	_ = azurePeeringCreateCmd.MarkFlagRequired("subscription-id")
	azurePeeringCreateCmd.Flags().StringVar(&azureVnetName, "vnet", "", "name of the azure vnet")
	_ = azurePeeringCreateCmd.MarkFlagRequired("vnet")

}
