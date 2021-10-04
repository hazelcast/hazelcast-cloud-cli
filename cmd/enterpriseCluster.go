package cmd

import (
	"context"
	"github.com/fatih/color"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"github.com/jedib0t/go-pretty/v6/table"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var enterpriseClusterId string
var enterpriseClusterCreateInput models.CreateEnterpriseClusterInput
var enterpriseClusterCreateZoneType string

var enterpriseClusterCmd = &cobra.Command{
	Use:     "enterprise-cluster",
	Aliases: []string{"ec"},
	Short:   "This command allows you to make actions on your enterprise clusters like; create and delete.",
}

var enterpriseClusterCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "This command creates Hazelcast instance with provided configurations.",
	Example: "hzcloud enterprise-cluster create --name=mycluster2 --cloud-provider=aws --region=eu-west-2 --hazelcast-version=4.0 --instance-type=m5.large --cidr-block=10.0.80.0/16 --native-memory=4 --wait",
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneType, err := util.AugmentZoneType(enterpriseClusterCreateZoneType)
		if err != nil {
			return err
		}
		enterpriseClusterCreateInput.ZoneType = zoneType
		client := internal.NewClient()
		cluster := internal.Validate(client.EnterpriseCluster.Create(context.Background(), &enterpriseClusterCreateInput)).(*models.Cluster)
		color.Green("Cluster creation started.")
		waitCondition, _ := cmd.Flags().GetBool("wait")
		if waitCondition {
			loading := util.NewLoadingIndicator("Waiting", 4)
			loading.Start()
			for {
				cluster, _, _ = client.EnterpriseCluster.Get(context.Background(), &models.GetEnterpriseClusterInput{ClusterId: cluster.Id})
				loading.SetStep(cluster.Progress.Status, cluster.Progress.CompletedItemCount)
				if cluster.State == models.Running {
					loading.Stop()
					color.Green("Cluster successfully created in %s", loading.Stop())
					break
				} else if cluster.State == models.Failed {
					loading.Stop()
					color.Red("Cluster failed in %s", loading.Stop())
					break
				}
				time.Sleep(5 * time.Second)
			}
		}
		return nil
	},
}

var enterpriseClusterGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "This command get detailed configuration of starter Hazelcast instance.",
	Example: "hzcloud enterprise-cluster get --cluster-id=3",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		cluster := internal.Validate(client.EnterpriseCluster.Get(context.Background(), &models.GetEnterpriseClusterInput{
			ClusterId: enterpriseClusterId,
		})).(*models.Cluster)
		util.Print(util.PrintRequest{
			Data:       *cluster,
			PrintStyle: util.PrintStyle(outputStyle),
		})
	},
}

var enterpriseClusterListCmd = &cobra.Command{
	Use:     "list",
	Short:   "This command lists Hazelcast Instances.",
	Example: "hzcloud enterprise-cluster list",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		clusters := internal.Validate(client.EnterpriseCluster.List(context.Background())).(*[]models.Cluster)
		header := table.Row{"Id", "Name", "State", "Version", "Memory(GiB)", "Network", "Instance", "Per Zone", "Cloud Provider", "Region", "Zones"}
		rows := []table.Row{}
		for _, cluster := range *clusters {
			rows = append(rows, table.Row{cluster.Id, cluster.Name, cluster.State, cluster.HazelcastVersion,
				cluster.Specs.TotalMemory, cluster.Networking.Type, cluster.Specs.InstanceType, cluster.Specs.InstancePerZone,
				cluster.CloudProvider.Name, cluster.CloudProvider.Region,
				strings.Join(cluster.CloudProvider.AvailabilityZones, ", ")})
		}
		util.Print(util.PrintRequest{
			Header:     header,
			Rows:       rows,
			Data:       clusters,
			PrintStyle: util.PrintStyle(outputStyle),
		})
	},
}

var enterpriseClusterDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "This command deletes Hazelcast Instance according to its id",
	Example: "hzcloud enterprise-cluster delete --cluster-id=3",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		clusterResponse := internal.Validate(client.EnterpriseCluster.Delete(context.Background(), &models.ClusterDeleteInput{
			ClusterId: enterpriseClusterId,
		})).(*models.ClusterId)
		color.Blue("Cluster %d deleted.", clusterResponse.ClusterId)
	},
}

func init() {
	rootCmd.AddCommand(enterpriseClusterCmd)
	enterpriseClusterCmd.AddCommand(enterpriseClusterCreateCmd)
	enterpriseClusterCmd.AddCommand(enterpriseClusterGetCmd)
	enterpriseClusterCmd.AddCommand(enterpriseClusterListCmd)
	enterpriseClusterCmd.AddCommand(enterpriseClusterDeleteCmd)

	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateInput.Name, "name", "", "name of the cluster")
	err := enterpriseClusterCreateCmd.MarkFlagRequired("name")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateInput.CloudProvider, "cloud-provider", "", "name of the cloud provider")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("cloud-provider")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateInput.Region, "region", "", "name of the region")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("region")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateInput.HazelcastVersion, "hazelcast-version", "", "version of the Hazelcast")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("hazelcast-version")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateInput.InstanceType, "instance-type", "", "physical type of the instance")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("instance-type")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateZoneType, "zone-type", "", "zone type of the cluster")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("zone-type")

	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateInput.CidrBlock, "cidr-block", "", "cidr block of the instance")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("cidr-block")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().IntVar(&enterpriseClusterCreateInput.NativeMemory, "native-memory", 0, "native memory as gigabyte")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("native-memory")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().IntVar(&enterpriseClusterCreateInput.InstancePerZone, "instance-per-zone", 1, "instance per zone")
	enterpriseClusterCreateCmd.Flags().BoolVar(&enterpriseClusterCreateInput.IsPublicAccessEnabled, "public-access-enabled", false, "public access")
	enterpriseClusterCreateCmd.Flags().BoolVar(&enterpriseClusterCreateInput.IsAutoScalingEnabled, "auto-scaling-enabled", false, "auto scaling feature")
	enterpriseClusterCreateCmd.Flags().BoolVar(&enterpriseClusterCreateInput.IsHotRestartEnabled, "hot-restart-enabled", false, "hot restart feature")
	enterpriseClusterCreateCmd.Flags().BoolVar(&enterpriseClusterCreateInput.IsHotBackupEnabled, "hot-backup-enabled", false, "hot backup feature")
	enterpriseClusterCreateCmd.Flags().BoolVar(&enterpriseClusterCreateInput.IsTLSEnabled, "tls-enabled", false, "tls encryption feature")
	enterpriseClusterCreateCmd.Flags().Bool("wait", false, "wait until resource creation finished")

	enterpriseClusterDeleteCmd.Flags().StringVar(&enterpriseClusterId, "cluster-id", "", "id of the cluster")
	err = enterpriseClusterDeleteCmd.MarkFlagRequired("cluster-id")
	if err != nil {
		panic(err)
	}

	enterpriseClusterGetCmd.Flags().StringVar(&enterpriseClusterId, "cluster-id", "", "id of the cluster")
	err = enterpriseClusterGetCmd.MarkFlagRequired("cluster-id")
	if err != nil {
		panic(err)
	}
}
