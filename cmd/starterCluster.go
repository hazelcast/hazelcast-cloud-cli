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

var starterClusterId string

var starterClusterCreateInput models.CreateStarterClusterInput

var starterClusterCreateClusterType string
var starterClusterCreateHazelcastVersion float64

var starterClusterCmd = &cobra.Command{
	Use:     "starter-cluster",
	Aliases: []string{"sc"},
	Short:   "This command allows you to make actions on your starter clusters like; create, delete, stop or resume.",
}

var starterClusterCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "This command creates Hazelcast instance with provided configurations.",
	Example: "hzcloud starter-cluster create --cloud-provider=aws --cluster-type=FREE --name=mycluster --region=us-west-2 --total-memory=0.2 --hazelcast-version=4.0",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := internal.NewClient()
		clusterType, err := util.AugmentStarterClusterType(starterClusterCreateClusterType)
		if err != nil {
			return err
		}

		hazelcastVersion, err2 := util.AugmentStarterHazelcastVersion(starterClusterCreateHazelcastVersion)
		if err2 != nil {
			return err2
		}

		starterClusterCreateInput.ClusterType = clusterType
		starterClusterCreateInput.HazelcastVersion = hazelcastVersion
		cluster := internal.Validate(client.StarterCluster.Create(context.Background(), &starterClusterCreateInput)).(*models.Cluster)
		color.Green("Cluster %s is creating.", cluster.Id)
		return nil
	},
}

var starterClusterGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "This command get detailed configuration of starter Hazelcast instance.",
	Example: "hzcloud starter-cluster get --cluster-id=100",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		cluster := internal.Validate(client.StarterCluster.Get(context.Background(), &models.GetStarterClusterInput{
			ClusterId: starterClusterId,
		})).(*models.Cluster)
		util.Print(util.PrintRequest{
			Data:       *cluster,
			PrintStyle: util.PrintStyle(outputStyle),
		})
	},
}

var starterClusterListCmd = &cobra.Command{
	Use:     "list",
	Short:   "This command lists Hazelcast Instances.",
	Example: "hzcloud starter-cluster list",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		clusters := internal.Validate(client.StarterCluster.List(context.Background())).(*[]models.Cluster)
		header := table.Row{"Id", "Name", "State", "Version", "Memory (GiB)", "Cloud Provider", "Region", "Is Free"}
		rows := []table.Row{}
		for _, cluster := range *clusters {
			rows = append(rows, table.Row{cluster.Id, cluster.Name, cluster.State, cluster.HazelcastVersion,
				cluster.Specs.TotalMemory, cluster.CloudProvider.Name, cluster.CloudProvider.Region, cluster.ProductType.IsFree})
		}
		util.Print(util.PrintRequest{
			Header:     header,
			Rows:       rows,
			Data:       clusters,
			PrintStyle: util.PrintStyle(outputStyle),
		})
	},
}

var starterClusterDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "This command deletes Hazelcast Instance according to its id",
	Example: "hzcloud starter-cluster delete --cluster-id=100",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		clusterResponse := internal.Validate(client.StarterCluster.Delete(context.Background(), &models.ClusterDeleteInput{
			ClusterId: starterClusterId,
		})).(*models.ClusterId)
		color.Blue("Cluster %d deleted.", clusterResponse.ClusterId)
	},
}

var starterClusterStopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "This command stops Hazelcast Instance according to its id",
	Example: "hzcloud starter-cluster stop --cluster-id=100",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		clusterResponse := internal.Validate(client.StarterCluster.Stop(context.Background(), &models.ClusterStopInput{
			ClusterId: starterClusterId,
		})).(*models.ClusterId)
		color.Blue("Cluster %d stopped.", clusterResponse.ClusterId)
	},
}

var starterClusterResumeCmd = &cobra.Command{
	Use:     "resume",
	Short:   "This command resumes Hazelcast Instance according to its id",
	Example: "hzcloud starter-cluster resume --cluster-id=100",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		clusterResponse := internal.Validate(client.StarterCluster.Resume(context.Background(), &models.ClusterResumeInput{
			ClusterId: starterClusterId,
		})).(*models.ClusterId)
		color.Blue("Cluster %d resumed.", clusterResponse.ClusterId)
	},
}

func init() {
	rootCmd.AddCommand(starterClusterCmd)
	starterClusterCmd.AddCommand(starterClusterListCmd)
	starterClusterCmd.AddCommand(starterClusterGetCmd)
	starterClusterGetCmd.Flags().StringVar(&starterClusterId, "cluster-id", "", "id of the cluster")
	_ = starterClusterGetCmd.MarkFlagRequired("cluster-id")

	starterClusterCmd.AddCommand(starterClusterCreateCmd)
	starterClusterCreateCmd.Flags().StringVar(&starterClusterCreateInput.Name, "name", "", "name of the cluster")
	_ = starterClusterCreateCmd.MarkFlagRequired("name")
	starterClusterCreateCmd.Flags().StringVar(&starterClusterCreateInput.CloudProvider, "cloud-provider", "", "name of the cloud provider")
	_ = starterClusterCreateCmd.MarkFlagRequired("cloud-provider")
	starterClusterCreateCmd.Flags().StringVar(&starterClusterCreateInput.Region, "region", "", "name of the region")
	_ = starterClusterCreateCmd.MarkFlagRequired("region")
	starterClusterCreateCmd.Flags().StringVar(&starterClusterCreateClusterType, "cluster-type", "", "type of the cluster")
	_ = starterClusterCreateCmd.MarkFlagRequired("cluster-type")
	starterClusterCreateCmd.Flags().Float64Var(&starterClusterCreateInput.TotalMemory, "total-memory", 0, "total memory of cluster as gb")
	_ = starterClusterCreateCmd.MarkFlagRequired("total-memory")
	starterClusterCreateCmd.Flags().Float64Var(&starterClusterCreateHazelcastVersion, "hazelcast-version", 0, "version of hazelcast")
	_ = starterClusterCreateCmd.MarkFlagRequired("hazelcast-version")

	starterClusterCreateCmd.Flags().BoolVar(&starterClusterCreateInput.IsAutoScalingEnabled, "auto-scaling-enabled", false, "auto scaling feature")
	starterClusterCreateCmd.Flags().BoolVar(&starterClusterCreateInput.IsHotBackupEnabled, "hot-backup-enabled", false, "hot backup feature")
	starterClusterCreateCmd.Flags().BoolVar(&starterClusterCreateInput.IsHotRestartEnabled, "hot-restart-enabled", false, "hot restart feature")
	starterClusterCreateCmd.Flags().BoolVar(&starterClusterCreateInput.IsIPWhitelistEnabled, "ip-whitelist-enabled", false, "ip whitelist feature")
	starterClusterCreateCmd.Flags().StringSliceVar(&starterClusterCreateInput.IPWhitelist, "ip-whitelist", []string{}, "ip whitelist of cluster")

	starterClusterCmd.AddCommand(starterClusterDeleteCmd)
	starterClusterDeleteCmd.Flags().StringVar(&starterClusterId, "cluster-id", "", "id of the cluster")
	_ = starterClusterDeleteCmd.MarkFlagRequired("cluster-id")

	starterClusterCmd.AddCommand(starterClusterStopCmd)
	starterClusterStopCmd.Flags().StringVar(&starterClusterId, "cluster-id", "", "id of the cluster")
	_ = starterClusterStopCmd.MarkFlagRequired("cluster-id")

	starterClusterCmd.AddCommand(starterClusterResumeCmd)
	starterClusterResumeCmd.Flags().StringVar(&starterClusterId, "cluster-id", "", "id of the cluster")
	_ = starterClusterResumeCmd.MarkFlagRequired("cluster-id")

}
