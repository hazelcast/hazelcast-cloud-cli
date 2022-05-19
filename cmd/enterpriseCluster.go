package cmd

import (
	"context"
	"github.com/fatih/color"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/schollz/progressbar/v3"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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
	Example: "hzcloud enterprise-cluster create --name=mycluster2 --cloud-provider=aws --region=eu-west-2 --zone-type=SINGLE --hazelcast-version=4.0 --instance-type=m5.large --cidr-block=10.80.0.0/16 --native-memory=4 --wait",
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneType, err := util.AugmentZoneType(enterpriseClusterCreateZoneType)
		if err != nil {
			return err
		}
		enterpriseClusterCreateInput.ZoneType = zoneType
		client := internal.NewClient()
		cluster := internal.Validate(client.EnterpriseCluster.Create(context.Background(),
			&enterpriseClusterCreateInput)).(*models.Cluster)
		color.Green("Cluster creation started.")
		waitCondition, _ := cmd.Flags().GetBool("wait")
		if waitCondition {
			loading := util.NewLoadingIndicator("Waiting", 4)
			loading.Start()
			for {
				cluster, _, _ = client.EnterpriseCluster.Get(context.Background(),
					&models.GetEnterpriseClusterInput{ClusterId: cluster.Id})
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
		cluster := internal.Validate(client.EnterpriseCluster.Get(context.Background(),
			&models.GetEnterpriseClusterInput{
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
		header := table.Row{
			"Id", "Name", "State", "Version", "Memory(GiB)", "Network", "Instance", "Per Zone", "Cloud Provider",
			"Region", "Zones",
		}
		rows := []table.Row{}
		for _, cluster := range *clusters {
			rows = append(rows, table.Row{
				cluster.Id, cluster.Name, cluster.State, cluster.HazelcastVersion,
				cluster.Specs.TotalMemory, cluster.Networking.Type, cluster.Specs.InstanceType,
				cluster.Specs.InstancePerZone,
				cluster.CloudProvider.Name, cluster.CloudProvider.Region,
				strings.Join(cluster.CloudProvider.AvailabilityZones, ", "),
			})
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
		clusterResponse := internal.Validate(client.EnterpriseCluster.Delete(context.Background(),
			&models.ClusterDeleteInput{
				ClusterId: enterpriseClusterId,
			})).(*models.ClusterId)
		color.Blue("Cluster %d deleted.", clusterResponse.ClusterId)
	},
}

func newEnterpriseCustomClassesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "custom-classes",
		Aliases: []string{"class"},
		Short:   "This command allows you to manage custom classes on your enterprise cluster like: list, upload, delete.",
	}
}

func newEnterpriseCustomClassesListCmd() *cobra.Command {
	var clusterId string

	enterpriseCustomClassesListCmd := cobra.Command{
		Use:     "list",
		Short:   "This command lists Artifacts that contains Custom Classes uploaded to Hazelcast Instance.",
		Example: "hzcloud enterprise-cluster custom-classes list",
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.NewClient()
			artifacts := internal.Validate(client.EnterpriseCluster.ListUploadedArtifacts(context.Background(),
				&models.ListUploadedArtifactsInput{
					ClusterId: clusterId,
				})).(*[]models.UploadedArtifact)
			header := table.Row{"Id", "File Name", "Status"}
			rows := []table.Row{}
			for _, artifact := range *artifacts {
				rows = append(rows, table.Row{artifact.Id, artifact.Name, artifact.Status})
			}
			util.Print(util.PrintRequest{
				Header:     header,
				Rows:       rows,
				Data:       artifacts,
				PrintStyle: util.PrintStyle(outputStyle),
			})
		},
	}

	enterpriseCustomClassesListCmd.Flags().StringVar(&clusterId, "cluster-id", "", "id of the cluster")
	_ = enterpriseCustomClassesListCmd.MarkFlagRequired("cluster-id")

	return &enterpriseCustomClassesListCmd
}

func newEnterpriseClusterCustomClassesUploadCmd() *cobra.Command {
	var clusterId string
	var customClassesFileName string

	enterpriseClusterCustomClassesUploadCmd := cobra.Command{
		Use:     "upload",
		Short:   "This command uploads Artifact with custom classes to Hazelcast Instance.",
		Example: "hzcloud enterprise-cluster custom-classes upload",
		Run: func(cmd *cobra.Command, args []string) {
			file, err := os.Open(customClassesFileName)
			if err != nil {
				color.Red(err.Error())
				os.Exit(1)
			}
			defer file.Close()

			client := internal.NewClient()
			artifact := internal.Validate(client.EnterpriseCluster.UploadArtifact(context.Background(),
				&models.UploadArtifactInput{
					ClusterId: clusterId,
					FileName:  filepath.Base(file.Name()),
					Content:   file,
				})).(*models.UploadedArtifact)

			header := table.Row{"Id", "File Name", "Status"}
			rows := []table.Row{{artifact.Id, artifact.Name, artifact.Status}}
			util.Print(util.PrintRequest{
				Header:     header,
				Rows:       rows,
				Data:       artifact,
				PrintStyle: util.PrintStyle(outputStyle),
			})
		},
	}

	enterpriseClusterCustomClassesUploadCmd.Flags().StringVar(&clusterId, "cluster-id", "", "id of the cluster")
	enterpriseClusterCustomClassesUploadCmd.Flags().StringVar(&customClassesFileName, "file-name", "", "File to upload")
	_ = enterpriseClusterCustomClassesUploadCmd.MarkFlagRequired("cluster-id")
	_ = enterpriseClusterCustomClassesUploadCmd.MarkFlagRequired("file-name")

	return &enterpriseClusterCustomClassesUploadCmd
}

func newEnterpriseClusterCustomClassesDeleteCmd() *cobra.Command {
	var clusterId string
	var customClassesId string

	enterpriseClusterCustomClassesDeleteCmd := cobra.Command{
		Use:     "delete",
		Short:   "This command deletes Artifact with custom classes that was uploaded to Hazelcast Instance.",
		Example: "hzcloud enterprise-cluster custom-classes delete",
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.NewClient()
			artifact := internal.Validate(client.EnterpriseCluster.DeleteArtifact(context.Background(),
				&models.DeleteArtifactInput{
					ClusterId:       clusterId,
					CustomClassesId: customClassesId,
				})).(*models.UploadedArtifact)

			header := table.Row{"Id", "File Name", "Status"}
			rows := []table.Row{{artifact.Id, artifact.Name, artifact.Status}}
			util.Print(util.PrintRequest{
				Header:     header,
				Rows:       rows,
				Data:       artifact,
				PrintStyle: util.PrintStyle(outputStyle),
			})
		},
	}

	enterpriseClusterCustomClassesDeleteCmd.Flags().StringVar(&clusterId, "cluster-id", "", "id of the cluster")
	enterpriseClusterCustomClassesDeleteCmd.Flags().StringVar(&customClassesId, "file-id", "",
		"id of the Uploaded Artifact")
	_ = enterpriseClusterCustomClassesDeleteCmd.MarkFlagRequired("cluster-id")
	_ = enterpriseClusterCustomClassesDeleteCmd.MarkFlagRequired("file-id")

	return &enterpriseClusterCustomClassesDeleteCmd
}

func newEnterpriseClusterCustomClassesDownloadCmd() *cobra.Command {
	var clusterId string
	var customClassesId string

	enterpriseClusterCustomClassesDownloadCmd := cobra.Command{
		Use:     "download",
		Short:   "This command downloads an artifact with custom classes that was uploaded to Hazelcast Instance.",
		Example: "hzcloud enterprise-cluster custom-classes download",
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.NewClient()
			artifact := internal.Validate(client.EnterpriseCluster.DownloadArtifact(context.Background(),
				&models.DownloadArtifactInput{
					ClusterId:       clusterId,
					CustomClassesId: customClassesId,
				})).(*models.UploadedArtifactLink)

			header := table.Row{"Id", "File Name"}
			rows := []table.Row{{artifact.Id, artifact.Name}}
			util.Print(util.PrintRequest{
				Header:     header,
				Rows:       rows,
				Data:       artifact,
				PrintStyle: util.PrintStyle(outputStyle),
			})
			httpClient := http.Client{}
			request, requestErr := http.NewRequest("GET", artifact.Url, nil)
			if requestErr != nil {
				color.Red(requestErr.Error())
				os.Exit(1)
			}
			response, responseErr := httpClient.Do(request)
			if responseErr != nil {
				color.Red(responseErr.Error())
				os.Exit(1)
			}
			tmpFile, tmpFileErr := ioutil.TempFile("", "*-"+artifact.Name)
			if tmpFileErr != nil {
				color.Red(tmpFileErr.Error())
				os.Exit(1)
			}
			defer tmpFile.Close()
			bar := progressbar.DefaultBytes(
				response.ContentLength,
				"downloading "+artifact.Name,
			)
			_, writeErr := io.Copy(io.MultiWriter(tmpFile, bar), response.Body)
			if writeErr != nil {
				color.Red(writeErr.Error())
				os.Exit(1)
			}
			renameErr := os.Rename(tmpFile.Name(), artifact.Name)
			if renameErr != nil {
				color.Red(renameErr.Error())
				os.Exit(1)
			}
		},
	}

	enterpriseClusterCustomClassesDownloadCmd.Flags().StringVar(&clusterId, "cluster-id", "", "id of the cluster")
	enterpriseClusterCustomClassesDownloadCmd.Flags().StringVar(&customClassesId, "file-id", "",
		"id of the Uploaded Artifact")
	_ = enterpriseClusterCustomClassesDownloadCmd.MarkFlagRequired("cluster-id")
	_ = enterpriseClusterCustomClassesDownloadCmd.MarkFlagRequired("file-id")

	return &enterpriseClusterCustomClassesDownloadCmd
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
	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateInput.CloudProvider, "cloud-provider", "",
		"name of the cloud provider")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("cloud-provider")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateInput.Region, "region", "",
		"name of the region")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("region")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateInput.HazelcastVersion, "hazelcast-version",
		"", "version of the Hazelcast")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("hazelcast-version")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateInput.InstanceType, "instance-type", "",
		"physical type of the instance")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("instance-type")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateZoneType, "zone-type", "",
		"zone type of the cluster")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("zone-type")

	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().StringVar(&enterpriseClusterCreateInput.CidrBlock, "cidr-block", "",
		"cidr block of the instance")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("cidr-block")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().IntVar(&enterpriseClusterCreateInput.NativeMemory, "native-memory", 0,
		"native memory as gigabyte")
	err = enterpriseClusterCreateCmd.MarkFlagRequired("native-memory")
	if err != nil {
		panic(err)
	}
	enterpriseClusterCreateCmd.Flags().IntVar(&enterpriseClusterCreateInput.InstancePerZone, "instance-per-zone", 1,
		"instance per zone")
	enterpriseClusterCreateCmd.Flags().BoolVar(&enterpriseClusterCreateInput.IsPublicAccessEnabled,
		"public-access-enabled", false, "public access")
	enterpriseClusterCreateCmd.Flags().BoolVar(&enterpriseClusterCreateInput.IsAutoScalingEnabled,
		"auto-scaling-enabled", false, "auto scaling feature")
	enterpriseClusterCreateCmd.Flags().BoolVar(&enterpriseClusterCreateInput.IsHotRestartEnabled, "hot-restart-enabled",
		false, "hot restart feature")
	enterpriseClusterCreateCmd.Flags().BoolVar(&enterpriseClusterCreateInput.IsHotBackupEnabled, "hot-backup-enabled",
		false, "hot backup feature")
	enterpriseClusterCreateCmd.Flags().BoolVar(&enterpriseClusterCreateInput.IsTLSEnabled, "tls-enabled", false,
		"tls encryption feature")
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

	enterpriseCustomClassesCmd := newEnterpriseCustomClassesCmd()
	enterpriseClusterCmd.AddCommand(enterpriseCustomClassesCmd)
	enterpriseCustomClassesCmd.AddCommand(newEnterpriseCustomClassesListCmd())
	enterpriseCustomClassesCmd.AddCommand(newEnterpriseClusterCustomClassesUploadCmd())
	enterpriseCustomClassesCmd.AddCommand(newEnterpriseClusterCustomClassesDeleteCmd())
	enterpriseCustomClassesCmd.AddCommand(newEnterpriseClusterCustomClassesDownloadCmd())
}
