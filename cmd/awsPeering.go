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

var awsPeeringId string
var awsRegion string
var awsVpcId string
var awsSubnetIds []string

var awsPeeringCmd = &cobra.Command{
	Use:     "aws-peering",
	Short:   "This command allows you to make Amazon Web Service VPC related actions.",
	Aliases: []string{"aws"},
}

var awsPeeringCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "This command creates AWS VPC Peering between your own vpc and your Enterprise Hazelcast cluster.",
	Example: "hzcloud aws-peering create --cluster-id=1 --vpc-id=2 --subnet-ids=a,b,c",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		indicator := util.NewLoadingIndicator("AWS Peering starting...", 100)
		indicator.Start()
		awsPeeringService := service.NewAwsPeeringService(client, &service.AwsCustomerPeeringProperties{
			ClusterId: enterpriseClusterId,
			Region:    awsRegion,
			VpcId:     awsVpcId,
			SubnetIds: awsSubnetIds,
		})
		peeringCreateErr := awsPeeringService.Create(indicator)
		indicator.Stop()
		if peeringCreateErr != nil {
			color.Red("An error occurred. %s", peeringCreateErr)
		} else {
			color.Green("Peering successfully established.")
		}
	},
}

var awsPeeringListCmd = &cobra.Command{
	Use:     "list",
	Short:   "This command lists AWS VPC peerings on your Enterprise Hazelcast cluster.",
	Example: "hzcloud aws-peering list --cluster-id=1",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		peerings := internal.Validate(client.AwsPeering.List(context.Background(), &models.ListAwsPeeringsInput{
			ClusterId: enterpriseClusterId,
		})).(*[]models.AwsPeering)
		header := table.Row{"#", "Peering Id", "Vpc Id", "Vpc Cidr", "Subnet Id", "Subnet Cidr"}
		rows := []table.Row{}
		for k, peering := range *peerings {
			rows = append(rows, table.Row{k + 1, peering.Id, peering.VpcId, peering.VpcCidr, peering.SubnetId, peering.SubnetCidr})
		}
		util.Print(util.PrintRequest{
			Data:       peerings,
			Header:     header,
			Rows:       rows,
			PrintStyle: util.PrintStyle(outputStyle),
		})
	},
}

var awsPeeringDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "This command delete GCP peering from your Enterprise Hazelcast cluster.",
	Example: "hzcloud aws-peering delete --peering-id=1",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		_ = internal.Validate(client.AwsPeering.Delete(context.Background(), &models.DeleteAwsPeeringInput{
			Id: awsPeeringId,
		})).(*models.Result)
		color.Blue("Peering %s deleted.", awsPeeringId)
	},
}

func init() {
	rootCmd.AddCommand(awsPeeringCmd)
	awsPeeringCmd.AddCommand(awsPeeringCreateCmd)
	awsPeeringCmd.AddCommand(awsPeeringListCmd)
	awsPeeringCmd.AddCommand(awsPeeringDeleteCmd)

	awsPeeringListCmd.Flags().StringVar(&enterpriseClusterId, "cluster-id", "", "id of the cluster")
	_ = awsPeeringListCmd.MarkFlagRequired("cluster-id")

	awsPeeringDeleteCmd.Flags().StringVar(&awsPeeringId, "peering-id", "", "id of the peering")
	_ = awsPeeringDeleteCmd.MarkFlagRequired("peering-id")

	awsPeeringCreateCmd.Flags().StringVar(&enterpriseClusterId, "cluster-id", "", "id of the cluster")
	_ = awsPeeringCreateCmd.MarkFlagRequired("cluster-id")

	awsPeeringCreateCmd.Flags().StringVar(&awsRegion, "region", "", "id of the cluster")
	_ = awsPeeringCreateCmd.MarkFlagRequired("region")

	awsPeeringCreateCmd.Flags().StringVar(&awsVpcId, "vpc-id", "", "id of the cluster")
	_ = awsPeeringCreateCmd.MarkFlagRequired("vpc-id")

	awsPeeringCreateCmd.Flags().StringSliceVar(&awsSubnetIds, "subnet-ids", []string{}, "id of the cluster")
	_ = awsPeeringCreateCmd.MarkFlagRequired("subnet-ids")

}
