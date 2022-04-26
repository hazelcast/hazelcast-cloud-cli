package cmd

import (
	"context"
	"github.com/fatih/color"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newServerlessClusterCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "serverless-cluster",
		Aliases: []string{"slc"},
		Short:   "This command allows you to make actions on your serverless clusters like: create, delete, stop or resume.",
	}
}

func newServerlessClusterCreateCmd() *cobra.Command {
	var createClusterInputParams models.CreateServerlessClusterInput

	serverlessClusterCreateCmd := cobra.Command{
		Use:     "create",
		Short:   "This command allows you to create a serverless cluster.",
		Example: "hzcloud serverless-cluster create --name=mycluster --region=us-west-2",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := internal.NewClient()
			cluster := internal.Validate(client.ServerlessCluster.Create(context.Background(), &createClusterInputParams)).(*models.Cluster)
			color.Green("Cluster %s is creating. You can check the status using hzcloud serverless-cluster list.", cluster.Id)
			return nil
		},
	}

	serverlessClusterCreateCmd.Flags().StringVar(&createClusterInputParams.Name, "name", "", "name of the cluster (required)")
	_ = serverlessClusterCreateCmd.MarkFlagRequired("name")

	serverlessClusterCreateCmd.Flags().StringVar(&createClusterInputParams.Region, "region", "", "name of the region (required)")
	_ = serverlessClusterCreateCmd.MarkFlagRequired("region")

	var devModeEnabled bool
	serverlessClusterCreateCmd.Flags().BoolVar(&devModeEnabled, "dev-mode-enabled", false, "development mode")
	if devModeEnabled {
		createClusterInputParams.ClusterType = models.Devmode
	} else {
		createClusterInputParams.ClusterType = models.Serverless
	}
	return &serverlessClusterCreateCmd
}

func newServerlessClusterListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "This command allows you to create a serverless cluster.",
		Example: "hzcloud serverless-cluster list",
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.NewClient()
			clusters := internal.Validate(client.ServerlessCluster.List(context.Background())).(*[]models.Cluster)
			header := table.Row{"Id", "Name", "Type", "State", "Version", "Memory (GiB)", "Cloud Provider", "Region"}
			var rows []table.Row
			for _, cluster := range *clusters {
				rows = append(rows, table.Row{cluster.Id, cluster.Name, cluster.ClusterType.Name, cluster.State,
					cluster.HazelcastVersion, cluster.Specs.TotalMemory, cluster.CloudProvider.Name,
					cluster.CloudProvider.Region})
			}
			util.Print(util.PrintRequest{
				Header:     header,
				Rows:       rows,
				Data:       clusters,
				PrintStyle: util.PrintStyle(outputStyle),
			})
		},
	}
}

func newServerlessClusterGetCmd() *cobra.Command {
	var serverlessClusterId string

	serverlessClusterGetCmd := cobra.Command{
		Use:     "get",
		Short:   "This command get detailed configuration of serverless Hazelcast cluster instance.",
		Example: "hzcloud serverless-cluster get --cluster-id=100",
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.NewClient()
			cluster := internal.Validate(client.ServerlessCluster.Get(context.Background(), &models.GetServerlessClusterInput{
				ClusterId: serverlessClusterId,
			})).(*models.Cluster)
			util.Print(util.PrintRequest{
				Data:       *cluster,
				PrintStyle: util.PrintStyle(outputStyle),
			})
		},
	}

	serverlessClusterGetCmd.Flags().StringVar(&serverlessClusterId, "cluster-id", "", "id of the cluster")
	_ = serverlessClusterGetCmd.MarkFlagRequired("cluster-id")

	return &serverlessClusterGetCmd
}

func init() {
	serverlessClusterCmd := newServerlessClusterCmd()
	rootCmd.AddCommand(serverlessClusterCmd)

	serverlessClusterCmd.AddCommand(newServerlessClusterCreateCmd())
	serverlessClusterCmd.AddCommand(newServerlessClusterListCmd())
	serverlessClusterCmd.AddCommand(newServerlessClusterGetCmd())
}
