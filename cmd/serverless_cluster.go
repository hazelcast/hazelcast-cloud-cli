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
		Short:   "This command allows you to make actions on your serverless clusters like: create, delete, get, list, stop or resume.",
	}
}

func newServerlessClusterCreateCmd() *cobra.Command {
	var createClusterInputParams models.CreateServerlessClusterInput
	var devModeEnabled bool

	serverlessClusterCreateCmd := cobra.Command{
		Use:     "create",
		Short:   "This command allows you to create a serverless Hazelcast cluster.",
		Example: "hzcloud serverless-cluster create --name=mycluster --region=us-west-2",
		RunE: func(cmd *cobra.Command, args []string) error {
			if devModeEnabled {
				createClusterInputParams.ClusterType = models.Devmode
			} else {
				createClusterInputParams.ClusterType = models.Serverless
			}
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

	serverlessClusterCreateCmd.Flags().BoolVar(&devModeEnabled, "dev-mode-enabled", false, "development mode")

	return &serverlessClusterCreateCmd
}

func newServerlessClusterListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "This command allows you to create a serverless Hazelcast cluster.",
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
		Short:   "This command get detailed configuration of a serverless Hazelcast cluster.",
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

func newServerlessClusterDeleteCmd() *cobra.Command {
	var serverlessClusterId string

	serverlessClusterDeleteCmd := &cobra.Command{
		Use:     "delete",
		Short:   "This command allows you to delete a serverless Hazelcast cluster.",
		Example: "hzcloud serverless-cluster delete --cluster-id=100",
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.NewClient()
			clusterResponse := internal.Validate(client.ServerlessCluster.Delete(context.Background(), &models.ClusterDeleteInput{
				ClusterId: serverlessClusterId,
			})).(*models.ClusterId)
			color.Blue("Cluster %d deleted.", clusterResponse.ClusterId)
		},
	}

	serverlessClusterDeleteCmd.Flags().StringVar(&serverlessClusterId, "cluster-id", "", "id of the cluster")
	_ = serverlessClusterDeleteCmd.MarkFlagRequired("cluster-id")

	return serverlessClusterDeleteCmd
}

func newServerlessClusterStopCmd() *cobra.Command {
	var serverlessClusterId string

	serverlessClusterStopCmd := &cobra.Command{
		Use:     "stop",
		Short:   "This command allows you to stop a serverless Hazelcast cluster.",
		Example: "hzcloud serverless-cluster stop --cluster-id=100",
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.NewClient()
			clusterResponse := internal.Validate(client.ServerlessCluster.Stop(context.Background(), &models.ClusterStopInput{
				ClusterId: serverlessClusterId,
			})).(*models.ClusterId)
			color.Blue("Cluster %d stopped.", clusterResponse.ClusterId)
		},
	}

	serverlessClusterStopCmd.Flags().StringVar(&serverlessClusterId, "cluster-id", "", "id of the cluster")
	_ = serverlessClusterStopCmd.MarkFlagRequired("cluster-id")

	return serverlessClusterStopCmd
}

func newServerlessClusterResumeCmd() *cobra.Command {
	var serverlessClusterId string

	serverlessClusterResumeCmd := &cobra.Command{
		Use:     "resume",
		Short:   "This command allows you to resume a serverless Hazelcast cluster.",
		Example: "hzcloud serverless-cluster resume --cluster-id=100",
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.NewClient()
			clusterResponse := internal.Validate(client.ServerlessCluster.Resume(context.Background(), &models.ClusterResumeInput{
				ClusterId: serverlessClusterId,
			})).(*models.ClusterId)
			color.Blue("Cluster %d resumed.", clusterResponse.ClusterId)
		},
	}

	serverlessClusterResumeCmd.Flags().StringVar(&serverlessClusterId, "cluster-id", "", "id of the cluster")
	_ = serverlessClusterResumeCmd.MarkFlagRequired("cluster-id")

	return serverlessClusterResumeCmd
}

func init() {
	serverlessClusterCmd := newServerlessClusterCmd()
	rootCmd.AddCommand(serverlessClusterCmd)

	serverlessClusterCmd.AddCommand(newServerlessClusterCreateCmd())
	serverlessClusterCmd.AddCommand(newServerlessClusterListCmd())
	serverlessClusterCmd.AddCommand(newServerlessClusterGetCmd())
	serverlessClusterCmd.AddCommand(newServerlessClusterDeleteCmd())
	serverlessClusterCmd.AddCommand(newServerlessClusterStopCmd())
	serverlessClusterCmd.AddCommand(newServerlessClusterResumeCmd())
}
