package cmd

import (
	"context"
	"github.com/fatih/color"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"github.com/spf13/cobra"
)

var serverlessClusterCmd = &cobra.Command{
	Use:     "serverless-cluster",
	Aliases: []string{"slc"},
	Short:   "This command allows you to make actions on your serverless clusters like: create, delete, stop or resume.",
}

func getServerlessClusterCreateCmd() *cobra.Command {
	var createClusterInputParams models.CreateServerlessClusterInput

	serverlessClusterCreateCmd := cobra.Command{
		Use:   "create",
		Short: "hzcloud serverless-cluster create --name=mycluster --region=us-west-2",
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

	return &serverlessClusterCreateCmd
}

func init() {
	rootCmd.AddCommand(serverlessClusterCmd)
	serverlessClusterCmd.AddCommand(getServerlessClusterCreateCmd())
}
